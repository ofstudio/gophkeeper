package boltvault

import (
	"strings"

	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

// AttachmentPut adds a new attachment to the vault
func (v *BoltVault) AttachmentPut(meta *models.AttachmentMeta, data []byte) (*models.AttachmentMeta, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}

	if meta == nil || strings.TrimSpace(meta.FileName) == "" {
		return nil, repo.ErrInvalidArgument
	}

	// Check if attachment has id
	if len(meta.Id) == 0 {
		// Generate new id
		id, err := v.newID()
		if err != nil {
			return nil, err
		}
		meta.Id = id
	}

	// Check if attachment has CreatedAt
	now := nowFunc().Unix()
	if meta.CreatedAt == 0 {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now
	meta.FileSize = uint64(len(data))

	// Save attachment meta
	return meta, v.db.Update(func(tx *bbolt.Tx) error {
		return v.txAttachmentPut(tx, meta, data)
	})
}

// AttachmentDelete marks an attachment as deleted
func (v *BoltVault) AttachmentDelete(id []byte) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	return v.db.Update(func(tx *bbolt.Tx) error {
		// Check if attachment exists
		if err := v.txMessageGet(tx, bucketAttachmentsMeta, id, &models.AttachmentMeta{}); err != nil {
			return err
		}
		meta := &models.AttachmentMeta{Id: id, UpdatedAt: nowFunc().Unix(), Deleted: true}
		// mark attachment as deleted
		if err := v.txMessagePut(tx, bucketAttachmentsMeta, id, meta); err != nil {
			return err
		}
		// wipe attachment data
		wipedData, err := v.encryptData([]byte{})
		if err != nil {
			return err
		}
		return v.txValuePut(tx, bucketAttachmentsData, id, wipedData)
	})
}

// AttachmentMetaGet returns meta of an attachment
func (v *BoltVault) AttachmentMetaGet(id []byte) (*models.AttachmentMeta, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}

	meta := &models.AttachmentMeta{}
	return meta, v.db.View(func(tx *bbolt.Tx) error {
		return v.txMessageGet(tx, bucketAttachmentsMeta, id, meta)
	})
}

// AttachmentDataGet returns the data of an attachment
func (v *BoltVault) AttachmentDataGet(id []byte) ([]byte, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}
	var data []byte
	return data, v.db.View(func(tx *bbolt.Tx) error {
		encryptedData, err := v.txValueGet(tx, bucketAttachmentsData, id)
		if err != nil {
			return err
		}
		data, err = v.decryptData(encryptedData)
		return err
	})
}

// attachmentVacuum removes all attachments that have been marked as deleted
func (v *BoltVault) attachmentVacuum() error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucketMeta := tx.Bucket(bucketAttachmentsMeta)
		bucketData := tx.Bucket(bucketAttachmentsData)
		if bucketMeta == nil || bucketData == nil {
			return repo.ErrDBNotInitialized
		}

		// Iterate over all attachment meta
		return bucketMeta.ForEach(func(id, metaEncrypted []byte) error {
			meta := &models.AttachmentMeta{}
			if err := v.decryptMessage(metaEncrypted, meta); err != nil {
				return err
			}
			// Skip undeleted attachments
			if !meta.Deleted {
				return nil
			}
			// Delete attachment meta and data
			if err := bucketMeta.Delete(id); err != nil {
				return err
			}
			return bucketData.Delete(id)
		})
	})
}

// txAttachmentPut saves an attachment meta and data to the vault
func (v *BoltVault) txAttachmentPut(tx *bbolt.Tx, meta *models.AttachmentMeta, data []byte) error {
	// Save attachment meta
	if err := v.txMessagePut(tx, bucketAttachmentsMeta, meta.Id, meta); err != nil {
		return err
	}
	// Encrypt attachment data
	dataEncrypted, err := v.encryptData(data)
	if err != nil {
		return err
	}
	// Save attachment data
	return v.txValuePut(tx, bucketAttachmentsData, meta.Id, dataEncrypted)
}
