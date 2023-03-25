package boltvault

import (
	"time"

	"github.com/awnumar/memguard"
	"github.com/google/uuid"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/ofstudio/gophkeeper/internal/client/providers/encrypt"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

// vault buckets
type bucketName []byte

var (
	bucketItemsMeta       = bucketName("items_meta") //
	bucketItemsData       = bucketName("items_data") //
	bucketAttachmentsMeta = bucketName("attachments_meta")
	bucketAttachmentsData = bucketName("attachments_data")
	bucketSettings        = bucketName("settings")

	// vaultBuckets is a list of all buckets used by the vault
	vaultBuckets = []bucketName{
		bucketItemsMeta,
		bucketItemsData,
		bucketAttachmentsMeta,
		bucketAttachmentsData,
		bucketSettings,
	}
)

// BoltVault is a repo.Vault implementation for a BoltDB database
type BoltVault struct {
	db           *bbolt.DB
	masterKeyEnc *memguard.Enclave
	salt         []byte
	encryptor    encrypt.Provider
}

// NewBoltVault creates a new BoltVault instance
func NewBoltVault(filePath string) (*BoltVault, error) {
	db, err := bbolt.Open(filePath, 0600, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, repo.ErrDBFailedToOpen
	}
	v := &BoltVault{db: db}
	if err = v.migrate(); err != nil {
		return nil, err
	}
	return v, nil
}

// WithEncryptor sets the encryptor provider for the vault
func (v *BoltVault) WithEncryptor(encryptor encrypt.Provider) *BoltVault {
	v.encryptor = encryptor
	return v
}

// IsLocked returns true if the vault is locked
func (v *BoltVault) IsLocked() bool {
	return v.masterKeyEnc == nil || v.salt == nil
}

// Lock locks the vault
func (v *BoltVault) Lock() {
	v.masterKeyEnc = nil
	memguard.WipeBytes(v.salt)
	v.salt = nil
}

// Unlock unlocks the vault with the given master password
func (v *BoltVault) Unlock(masterPass []byte) error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	if v.encryptor == nil {
		return repo.ErrMissingEncryptProvider
	}
	return v.db.View(func(tx *bbolt.Tx) error {
		// Get keys from the vault
		keys, err := v.txVaultKeysGet(tx)
		if err != nil {
			return err
		}
		v.salt = keys.Salt
		// Decrypt the master key
		masterKey, err := v.encryptor.DecryptMasterKey(keys.MasterKeyEncrypted, masterPass, v.salt)
		if err != nil {
			return repo.ErrFailedToDecrypt
		}
		// Store the master key in an enclave (also wipes the original master key slice)
		v.masterKeyEnc = memguard.NewEnclave(masterKey)
		return nil
	})
}

// Vacuum removes all deleted items and attachments from the vault
func (v *BoltVault) Vacuum() error {
	if err := v.itemVacuum(); err != nil {
		return err
	}
	return v.attachmentVacuum()
}

// Purge removes all data from the vault
func (v *BoltVault) Purge() error {
	if v.db == nil {
		return repo.ErrDBNotInitialized
	}
	if err := v.db.Update(func(tx *bbolt.Tx) error {
		v.Lock()
		for _, b := range vaultBuckets {
			if err := tx.DeleteBucket(b); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return v.migrate()
}

// Close closes the vault
func (v *BoltVault) Close() error {
	v.Lock()
	if v.db == nil {
		return nil
	}
	if err := v.db.Close(); err != nil {
		return repo.ErrDBFailedToClose
	}
	v.db = nil
	return nil
}

// newID generates a new random UUID
func (v *BoltVault) newID() ([]byte, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, repo.ErrFailedToGenerateID
	}
	return []byte(id.String()), nil
}

// txMessagePut encrypts a proto.Message and saves it to specified bucket
func (v *BoltVault) txMessagePut(tx *bbolt.Tx, b bucketName, key []byte, m proto.Message) error {
	// Encrypt message
	encrypted, err := v.encryptMessage(m)
	if err != nil {
		return err
	}
	// Save encrypted to the vault
	return v.txValuePut(tx, b, key, encrypted)
}

// txMessageGet retrieves a proto.Message from the bucket and decrypts it
func (v *BoltVault) txMessageGet(tx *bbolt.Tx, b bucketName, key []byte, m proto.Message) error {
	encrypted, err := v.txValueGet(tx, b, key)
	if err != nil {
		return err
	}
	return v.decryptMessage(encrypted, m)
}

// txValuePut saves value to specified bucket
func (v *BoltVault) txValuePut(tx *bbolt.Tx, b bucketName, key []byte, val []byte) error {
	bucket := tx.Bucket(b)
	if bucket == nil {
		return repo.ErrDBNotInitialized
	}
	if err := bucket.Put(key, val); err != nil {
		return repo.ErrFailedToWrite
	}
	return nil
}

// txValueGet retrieves a value from specified bucket
func (v *BoltVault) txValueGet(tx *bbolt.Tx, b bucketName, key []byte) ([]byte, error) {
	bucket := tx.Bucket(b)
	if bucket == nil {
		return nil, repo.ErrDBNotInitialized
	}
	val := bucket.Get(key)
	if val == nil {
		return nil, repo.ErrNotFound
	}
	return val, nil
}

// encryptMessage marshals a proto.Message and encrypts it using the vault's master key
func (v *BoltVault) encryptMessage(m proto.Message) ([]byte, error) {
	b, err := proto.Marshal(m)
	if err != nil {
		return nil, repo.ErrFailedToMarshal
	}
	defer memguard.WipeBytes(b)
	return v.encryptData(b)
}

// decryptMessage decrypts a message using the vault's master key and un-marshals it into proto.Message
func (v *BoltVault) decryptMessage(b []byte, m proto.Message) error {
	var err error
	if b, err = v.decryptData(b); err != nil {
		return err
	}
	if err = proto.Unmarshal(b, m); err != nil {
		return repo.ErrFailedToUnmarshal
	}
	return nil
}

// encryptData encrypts a byte slice using the vault's master key
func (v *BoltVault) encryptData(data []byte) ([]byte, error) {
	if v.IsLocked() {
		return nil, repo.ErrLocked
	}
	if v.encryptor == nil {
		return nil, repo.ErrMissingEncryptProvider
	}

	masterKeyLB, err := v.masterKeyEnc.Open()
	if err != nil {
		return nil, repo.ErrFailedToOpenMasterKey
	}
	defer masterKeyLB.Destroy()

	data, err = v.encryptor.EncryptData(data, masterKeyLB.Bytes())
	if err != nil {
		return nil, repo.ErrFailedToEncrypt
	}
	return data, nil
}

// decryptData decrypts a byte slice using the vault's master key
func (v *BoltVault) decryptData(data []byte) ([]byte, error) {
	if v.IsLocked() {
		return nil, repo.ErrLocked
	}
	if v.encryptor == nil {
		return nil, repo.ErrMissingEncryptProvider
	}

	masterKeyLB, err := v.masterKeyEnc.Open()
	if err != nil {
		return nil, repo.ErrFailedToOpenMasterKey
	}
	defer masterKeyLB.Destroy()

	data, err = v.encryptor.DecryptData(data, masterKeyLB.Bytes())
	if err != nil {
		return nil, repo.ErrFailedToDecrypt
	}
	return data, nil
}

// nowFunc is used to mock time.Now in tests
var nowFunc = time.Now
