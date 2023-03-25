package repo

import (
	"github.com/ofstudio/gophkeeper/internal/client/models"
)

// Vault is the interface for the encrypted vault.
// It is used by the client to securely store and retrieve items, attachments,
// sync server configuration and key management.
type Vault interface {
	VaultItems
	VaultAttachments
	VaultSyncServer
	VaultKeys
	// IsLocked returns true if the vault is locked
	IsLocked() bool
	// Lock locks the vault
	Lock()
	// Unlock unlocks the vault with the given master password
	Unlock(masterPass []byte) error
	// Vacuum removes all deleted items and attachments from the vault
	Vacuum() error
	// Purge removes all data from the vault
	Purge() error
	// Close closes the vault
	Close() error
}

// VaultItems is the interface for managing items in the vault
type VaultItems interface {
	// ItemPut creates new item or updates an existing item in the vault
	ItemPut(meta *models.ItemMeta, data *models.ItemData) (*models.ItemMeta, error)
	// ItemMetaGet returns an item meta from the vault
	ItemMetaGet(id []byte) (*models.ItemMeta, error)
	// ItemDataGet returns the data of an item
	ItemDataGet(id []byte) (*models.ItemData, error)
	// ItemDelete marks an item as deleted
	ItemDelete(id []byte) error
	// ItemMetaList returns a list of all items meta in the vault.
	ItemMetaList() (models.ItemMetaList, error)
	// ItemMetaFilter returns a list of all items meta in the vault that match the filter
	ItemMetaFilter(filter string) (models.ItemMetaList, error)
}

// VaultAttachments is the interface for managing attachments in the vault
type VaultAttachments interface {
	// AttachmentPut adds a new attachment or updates an existing attachment in the vault
	AttachmentPut(meta *models.AttachmentMeta, data []byte) (*models.AttachmentMeta, error)
	// AttachmentMetaGet returns meta of an attachment
	AttachmentMetaGet(id []byte) (*models.AttachmentMeta, error)
	// AttachmentDataGet returns data of an attachment
	AttachmentDataGet(id []byte) ([]byte, error)
	// AttachmentDelete marks an attachment as deleted
	AttachmentDelete(id []byte) error
}

// VaultSyncServer is the interface for managing the sync server configuration
type VaultSyncServer interface {
	// SyncServerPut sets the sync server configuration
	SyncServerPut(srv *models.SyncServer) error
	// SyncServerGet returns the sync server configuration
	SyncServerGet() (*models.SyncServer, error)
	// SyncServerPurge removes the sync server configuration
	SyncServerPurge() error
}

// VaultKeys is the interface for managing the vault keys
type VaultKeys interface {
	// KeysGenerateNew generates new keys for the vault and encrypts them with the given master password
	KeysGenerateNew(masterPass []byte) error
	// KeysPut stores the vault keys
	KeysPut(masterPass []byte, keys *models.Keys) error
	// KeysGet returns the vault keys
	KeysGet() (*models.Keys, error)
	// KeysReEncrypt re-encrypts the keys with the new master password
	KeysReEncrypt(oldMasterPass, newMasterPass []byte) error
	// KeysExist returns true if the vault keys are available
	KeysExist() bool
}
