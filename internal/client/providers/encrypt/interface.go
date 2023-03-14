package encrypt

// Provider is the interface of the client encryption provider
type Provider interface {
	// NewKey generates a new key
	NewKey() ([]byte, error)
	// NewSalt generates a new salt
	NewSalt() ([]byte, error)
	// EncryptData encrypts the given data with the given master key
	EncryptData(data, key []byte) ([]byte, error)
	// DecryptData decrypts the given data with the given master key
	DecryptData(encryptedData, key []byte) ([]byte, error)
	// EncryptMasterKey encrypts the given master key with the given master password and salt
	EncryptMasterKey(masterKey, masterPass, salt []byte) ([]byte, error)
	// DecryptMasterKey decrypts the given master key with the given master password and salt
	DecryptMasterKey(encryptedMasterKey, masterPass, salt []byte) ([]byte, error)
}
