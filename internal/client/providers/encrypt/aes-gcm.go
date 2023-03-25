package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

// AESGCM is an encrypt.Provider implementation using aes-256-gcm
type AESGCM struct {
}

// NewAESGCM returns a new AESGCM instance
func NewAESGCM() *AESGCM {
	return &AESGCM{}
}

// NewKey generates a new 32 bytes long key using crypto/rand
func (p AESGCM) NewKey() ([]byte, error) {
	k := make([]byte, 32)
	n, err := rand.Read(k)
	if err != nil {
		return nil, err
	}
	if n != 32 {
		return nil, ErrFailedToGenerateKey
	}
	return k, nil
}

// NewSalt generates a new 32 bytes long salt using crypto/rand
func (p AESGCM) NewSalt() ([]byte, error) {
	return p.NewKey()
}

// EncryptData encrypts the given data with the given key using aes-256-gcm
func (p AESGCM) EncryptData(data, key []byte) ([]byte, error) {
	if data == nil {
		return nil, ErrNoData
	}
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	// get the gcm
	gcm, err := p.getGCM(key)
	if err != nil {
		return nil, ErrFailedToEncrypt
	}

	// create a new nonce. Nonce should be from GCM
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, ErrFailedToEncrypt
	}

	// encrypt the data
	encryptedData := gcm.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}

// DecryptData decrypts the given data with the given master key using aes-256-gcm
func (p AESGCM) DecryptData(encryptedData, key []byte) ([]byte, error) {
	if encryptedData == nil {
		return nil, ErrNoData
	}
	if len(key) != 32 {
		return nil, ErrInvalidKeyLength
	}

	// get the gcm
	gcm, err := p.getGCM(key)
	if err != nil {
		return nil, ErrFailedToDecrypt
	}
	nonceSize := gcm.NonceSize()

	// extract the nonce from the encrypted data
	if len(encryptedData) < nonceSize {
		return nil, ErrFailedToDecrypt
	}
	nonce, encryptedData := encryptedData[:nonceSize], encryptedData[nonceSize:]
	// decrypt the data
	data, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, ErrFailedToDecrypt
	}
	if data == nil {
		data = []byte{}
	}
	return data, nil
}

// EncryptMasterKey encrypts the given master key with the given master password and salt
// using aes-256-gcm and hmac-sha256
func (p AESGCM) EncryptMasterKey(masterKey, masterPass, salt []byte) ([]byte, error) {
	if len(masterPass) == 0 {
		return nil, ErrNoMasterPass
	}
	if len(masterKey) != 32 {
		return nil, ErrInvalidKeyLength
	}
	if len(salt) != 32 {
		return nil, ErrInvalidSaltLength
	}
	mac := hmac.New(sha256.New, salt)
	mac.Write(masterPass)
	key := mac.Sum(nil)
	return p.EncryptData(masterKey, key)
}

// DecryptMasterKey decrypts the given master key with the given master password and salt
// using aes-256-gcm and hmac-sha256
func (p AESGCM) DecryptMasterKey(encryptedMasterKey, masterPass, salt []byte) ([]byte, error) {
	if len(masterPass) == 0 {
		return nil, ErrNoMasterPass
	}
	if len(salt) != 32 {
		return nil, ErrInvalidSaltLength
	}
	mac := hmac.New(sha256.New, salt)
	mac.Write(masterPass)
	key := mac.Sum(nil)
	return p.DecryptData(encryptedMasterKey, key)
}

// getGCM returns a cipher.AEAD for the given master key
func (p AESGCM) getGCM(masterKey []byte) (cipher.AEAD, error) {
	// create a new aes cipher using the master key
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, err
	}
	// create a new GCM
	return cipher.NewGCM(block)
}
