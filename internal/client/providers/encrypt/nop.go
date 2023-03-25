package encrypt

import (
	"fmt"
	"os"
)

// Nop is a no-op encrypt.Provider.
// It does not encrypt or decrypt anything and returns a copies of input data instead.
// This is useful for testing and debugging purposes.
//
// * * *
//
// DO NOT USE THIS IN PRODUCTION!
//
// * * *
type Nop struct{}

func NewNop() *Nop {
	_, _ = fmt.Fprint(
		os.Stderr,
		"\n\n"+
			"WARNING: Using the Nop encryptor!\n"+
			"This is NOT SECURE and should only be used for testing purposes."+
			"\n\n",
	)
	return &Nop{}
}

func (p Nop) NewKey() ([]byte, error) {
	return []byte{0}, nil
}

func (p Nop) NewSalt() ([]byte, error) {
	return []byte{0}, nil
}

func (p Nop) EncryptData(data, _ []byte) ([]byte, error) {
	return append([]byte{}, data...), nil
}

func (p Nop) DecryptData(encryptedData, _ []byte) ([]byte, error) {
	return append([]byte{}, encryptedData...), nil
}

func (p Nop) EncryptMasterKey(masterKey, _, _ []byte) ([]byte, error) {
	return append([]byte{}, masterKey...), nil
}

func (p Nop) DecryptMasterKey(encryptedMasterKey, _, _ []byte) ([]byte, error) {
	return append([]byte{}, encryptedMasterKey...), nil
}
