package boltvault

import (
	"bytes"

	"github.com/awnumar/memguard"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

var keyVaultKeys = []byte("vault_keys")

// KeysGenerateNew generates new keys for the vault and encrypts them with the given master password
func (v *BoltVault) KeysGenerateNew(masterPass []byte) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	if v.encryptor == nil {
		return repo.ErrMissingEncryptProvider
	}
	// Check if keys already exist
	if v.KeysExist() {
		return repo.ErrAlreadyExists
	}
	// Lock the vault
	v.Lock()

	// Generate new keys
	salt, err := v.encryptor.NewSalt()
	if err != nil {
		return repo.ErrorFailedToGenerateSecret
	}
	masterKey, err := v.encryptor.NewKey()
	if err != nil {
		return repo.ErrorFailedToGenerateSecret
	}
	defer memguard.WipeBytes(masterKey)

	// Encrypt master key
	masterKeyEncrypted, err := v.encryptor.EncryptMasterKey(masterKey, masterPass, salt)
	if err != nil {
		return repo.ErrFailedToEncrypt
	}

	// Save keys
	return v.db.Update(func(tx *bbolt.Tx) error {
		return v.txVaultKeysPut(tx, &models.Keys{
			MasterKeyEncrypted: masterKeyEncrypted,
			Salt:               salt,
			UpdatedAt:          nowFunc().Unix(),
		})
	})
}

// KeysPut stores the vault keys.
// Checks if the provided master password is correct for the given keys.
// If vault keys already exist, checks if the new master key matches the current one.
func (v *BoltVault) KeysPut(masterPass []byte, keys *models.Keys) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	if v.encryptor == nil {
		return repo.ErrMissingEncryptProvider
	}

	// Decrypt new master key. Also checks if provided master password is correct.
	newMasterKey, err := v.encryptor.DecryptMasterKey(keys.MasterKeyEncrypted, masterPass, keys.Salt)
	if err != nil {
		return repo.ErrFailedToDecrypt
	}
	defer memguard.WipeBytes(newMasterKey)

	// If vault keys already exist,
	// check if new master key matches the current one
	if v.KeysExist() {
		if v.IsLocked() {
			return repo.ErrLocked
		}
		// Extract current master key
		masterKeyLB, err := v.masterKeyEnc.Open()
		if err != nil {
			return repo.ErrFailedToOpenMasterKey
		}
		defer masterKeyLB.Destroy()

		// Compare current master key with new master key
		if !bytes.Equal(masterKeyLB.Bytes(), newMasterKey) {
			return repo.ErrMasterKeyMismatch
		}
	}

	// Save new keys
	v.Lock()
	return v.db.Update(func(tx *bbolt.Tx) error {
		return v.txVaultKeysPut(tx, keys)
	})
}

// KeysGet returns the vault keys
func (v *BoltVault) KeysGet() (*models.Keys, error) {
	if v.db == nil {
		return nil, repo.ErrDBNotInitialized
	}
	var keys *models.Keys
	return keys, v.db.View(func(tx *bbolt.Tx) error {
		var err error
		keys, err = v.txVaultKeysGet(tx)
		return err
	})
}

// KeysReEncrypt re-encrypts the keys with the new master password
func (v *BoltVault) KeysReEncrypt(oldMasterPass, newMasterPass []byte) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	// Check if keys already exist
	if !v.KeysExist() {
		return repo.ErrNotFound
	}
	// Lock the vault
	v.Lock()

	return v.db.Update(func(tx *bbolt.Tx) error {
		// Get keys
		keys, err := v.txVaultKeysGet(tx)
		if err != nil {
			return err
		}
		// Decrypt master key
		masterKey, err := v.encryptor.DecryptMasterKey(keys.MasterKeyEncrypted, oldMasterPass, keys.Salt)
		if err != nil {
			return repo.ErrFailedToDecrypt
		}
		defer memguard.WipeBytes(masterKey)
		// Generate new salt
		salt, err := v.encryptor.NewSalt()
		if err != nil {
			return repo.ErrorFailedToGenerateSecret
		}
		// Encrypt master key with new salt and new master pass
		masterKeyEncrypted, err := v.encryptor.EncryptMasterKey(masterKey, newMasterPass, salt)
		if err != nil {
			return repo.ErrorFailedToGenerateSecret
		}
		// Save keys
		return v.txVaultKeysPut(tx, &models.Keys{
			MasterKeyEncrypted: masterKeyEncrypted,
			Salt:               salt,
			UpdatedAt:          nowFunc().Unix(),
		})
	})
}

// KeysExist returns true if the vault keys are available
func (v *BoltVault) KeysExist() bool {
	if v.db == nil {
		return false
	}
	var hasKeys bool
	_ = v.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketSettings)
		if bucket == nil {
			return repo.ErrDBNotInitialized
		}
		hasKeys = bucket.Get(keyVaultKeys) != nil
		return nil
	})
	return hasKeys
}

// txVaultKeysGet retrieves the vault keys
func (v *BoltVault) txVaultKeysGet(tx *bbolt.Tx) (*models.Keys, error) {
	// Get keys
	val, err := v.txValueGet(tx, bucketSettings, keyVaultKeys)
	if err != nil {
		return nil, err
	}
	keys := &models.Keys{}
	if err = proto.Unmarshal(val, keys); err != nil {
		return nil, repo.ErrFailedToUnmarshal
	}
	return keys, nil
}

// txVaultKeysPut saves the vault keys
func (v *BoltVault) txVaultKeysPut(tx *bbolt.Tx, keys *models.Keys) error {
	// Marshal keys
	val, err := proto.Marshal(keys)
	if err != nil {
		return err
	}
	// Save keys
	return v.txValuePut(tx, bucketSettings, keyVaultKeys, val)
}
