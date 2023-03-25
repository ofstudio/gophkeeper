package boltvault

import (
	"bytes"

	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

// keyDBVersion is the key of the database version
var keyDBVersion = []byte("db_version")

// dbCurrentVersion is the current version of the database
var dbCurrentVersion = []byte("1.0.0")

// migrate performs a migration of the vault database
func (v *BoltVault) migrate() error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}

	return v.db.Update(func(tx *bbolt.Tx) error {
		// Create buckets if they don't exist
		for _, bucket := range vaultBuckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return repo.ErrDBFailedToMigrate
			}
		}

		// Check the database version
		version, err := v.txDBVersionGet(tx, dbCurrentVersion)
		if err != nil {
			return repo.ErrDBFailedToMigrate
		}

		// We are currently supporting only one version 1.0.0
		if !bytes.Equal(version, dbCurrentVersion) {
			return repo.ErrDBVersionNotSupported
		}
		return nil
	})
}

// dbVersionGet returns the version of the database.
// If no version found, it sets the default version and returns it.
func (v *BoltVault) txDBVersionGet(tx *bbolt.Tx, defaultVer []byte) ([]byte, error) {
	version, err := v.txValueGet(tx, bucketSettings, keyDBVersion)
	if err != nil && err != repo.ErrNotFound {
		return nil, repo.ErrFailedToRead
	}
	if version == nil {
		version = append([]byte{}, defaultVer...)
		if err = v.txValuePut(tx, bucketSettings, keyDBVersion, version); err != nil {
			return nil, err
		}
	}
	return version, err
}
