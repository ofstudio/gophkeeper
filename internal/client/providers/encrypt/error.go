package encrypt

import "errors"

var (
	ErrInvalidKeyLength    = errors.New("invalid key length")
	ErrInvalidSaltLength   = errors.New("invalid salt length")
	ErrNoData              = errors.New("no data")
	ErrNoMasterPass        = errors.New("no master pass")
	ErrFailedToGenerateKey = errors.New("failed to generate key")
	ErrFailedToEncrypt     = errors.New("failed to encrypt")
	ErrFailedToDecrypt     = errors.New("failed to decrypt")
)
