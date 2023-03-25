package boltvault

import (
	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

var keySyncServer = []byte("sync_server")

// SyncServerGet returns the sync server configuration
func (v *BoltVault) SyncServerGet() (*models.SyncServer, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}
	srv := &models.SyncServer{}
	return srv, v.db.View(func(tx *bbolt.Tx) error {
		return v.txMessageGet(tx, bucketSettings, keySyncServer, srv)
	})
}

// SyncServerPut sets the sync server configuration
func (v *BoltVault) SyncServerPut(srv *models.SyncServer) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	return v.db.Update(func(tx *bbolt.Tx) error {
		return v.txMessagePut(tx, bucketSettings, keySyncServer, srv)
	})
}

// SyncServerPurge removes the sync server configuration
func (v *BoltVault) SyncServerPurge() error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketSettings)
		if bucket == nil {
			return repo.ErrDBNotInitialized
		}
		return bucket.Delete(keySyncServer)
	})
}
