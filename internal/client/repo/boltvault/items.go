package boltvault

import (
	"strings"

	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

// ItemPut creates new item or updates an existing item in the vault
func (v *BoltVault) ItemPut(meta *models.ItemMeta, data *models.ItemData) (*models.ItemMeta, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}
	if meta == nil || data == nil {
		return nil, repo.ErrInvalidArgument
	}

	// Check if item has id
	if len(meta.Id) == 0 {
		// Generate new id
		id, err := v.newID()
		if err != nil {
			return nil, err
		}
		meta.Id = id
	}

	// Check if item has CreatedAt
	now := nowFunc().Unix()
	if meta.CreatedAt == 0 {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now

	// Save item to the vault
	return meta, v.db.Update(func(tx *bbolt.Tx) error {
		return v.txItemPut(tx, meta, data)
	})
}

// ItemMetaGet returns an item meta from the vault
func (v *BoltVault) ItemMetaGet(id []byte) (*models.ItemMeta, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}

	meta := &models.ItemMeta{}
	return meta, v.db.View(func(tx *bbolt.Tx) error {
		return v.txMessageGet(tx, bucketItemsMeta, id, meta)
	})
}

// ItemDataGet returns the data of an item from the vault
func (v *BoltVault) ItemDataGet(id []byte) (*models.ItemData, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}

	data := &models.ItemData{}
	return data, v.db.View(func(tx *bbolt.Tx) error {
		return v.txMessageGet(tx, bucketItemsData, id, data)
	})
}

// ItemDelete marks an item as deleted
func (v *BoltVault) ItemDelete(id []byte) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	return v.db.Update(func(tx *bbolt.Tx) error {
		// check if item meta exists
		if err := v.txMessageGet(tx, bucketItemsMeta, id, &models.ItemMeta{}); err != nil {
			return err
		}
		// mark item as deleted
		meta := &models.ItemMeta{Id: id, UpdatedAt: nowFunc().Unix(), Deleted: true}
		data := &models.ItemData{}
		return v.txItemPut(tx, meta, data)
	})
}

// ItemMetaList returns a list of all items meta in the vault.
func (v *BoltVault) ItemMetaList() (models.ItemMetaList, error) {
	return v.ItemMetaFilter("")
}

// ItemMetaFilter returns a list of all items meta in the vault that match the filter
func (v *BoltVault) ItemMetaFilter(filter string) (models.ItemMetaList, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}

	var list models.ItemMetaList
	return list, v.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketItemsMeta)
		if bucket == nil {
			return repo.ErrDBNotInitialized
		}
		// Iterate over all list
		return bucket.ForEach(func(_, val []byte) error {
			meta := &models.ItemMeta{}
			if err := v.decryptMessage(val, meta); err != nil {
				return err
			}
			// Skip deleted items
			if meta.Deleted {
				return nil
			}
			// Add item to the list if it matches the filter or filter is empty
			if filter == "" || strings.Contains(meta.Title, filter) {
				list = append(list, meta)
			}
			return nil
		})
	})
}

// itemVacuum removes all items that have been marked as deleted
func (v *BoltVault) itemVacuum() error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}

	return v.db.Update(func(tx *bbolt.Tx) error {
		bucketMeta := tx.Bucket(bucketItemsMeta)
		bucketData := tx.Bucket(bucketItemsData)
		if bucketMeta == nil || bucketData == nil {
			return repo.ErrDBNotInitialized
		}
		// Iterate over all items
		return bucketMeta.ForEach(func(id, metaEncrypted []byte) error {
			meta := &models.ItemMeta{}
			if err := v.decryptMessage(metaEncrypted, meta); err != nil {
				return err
			}
			// Skip undeleted items
			if !meta.Deleted {
				return nil
			}
			// Delete item meta and data
			if err := bucketMeta.Delete(id); err != nil {
				return err
			}
			return bucketData.Delete(id)
		})
	})
}

// txItemPut saves an item meta and data to the vault
func (v *BoltVault) txItemPut(tx *bbolt.Tx, meta *models.ItemMeta, data *models.ItemData) error {
	if err := v.txMessagePut(tx, bucketItemsMeta, meta.Id, meta); err != nil {
		return err
	}
	return v.txMessagePut(tx, bucketItemsData, meta.Id, data)
}
