package repo

import "errors"

var (
	ErrDBNotInitialized      = errors.New("db not initialized")
	ErrDBFailedToOpen        = errors.New("failed to open db")
	ErrDBFailedToClose       = errors.New("failed to close db")
	ErrDBFailedToMigrate     = errors.New("failed to migrate db")
	ErrDBVersionNotSupported = errors.New("db version not supported")

	ErrLocked                   = errors.New("locked")
	ErrFailedToOpenMasterKey    = errors.New("failed to open master key")
	ErrFailedToGenerateID       = errors.New("failed to generate id")
	ErrFailedToMarshal          = errors.New("failed to marshal data")
	ErrFailedToUnmarshal        = errors.New("failed to unmarshal data")
	ErrMissingEncryptProvider   = errors.New("missing encrypt provider")
	ErrorFailedToGenerateSecret = errors.New("failed to generate secret")
	ErrFailedToEncrypt          = errors.New("failed to encrypt data")
	ErrFailedToDecrypt          = errors.New("failed to decrypt data")
	ErrMasterKeyMismatch        = errors.New("master key mismatch")

	ErrFailedToRead  = errors.New("failed to read data")
	ErrFailedToWrite = errors.New("failed to write data")

	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidArgument = errors.New("invalid argument")
)
