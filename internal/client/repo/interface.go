package repo

import (
	"io"

	"github.com/ofstudio/gophkeeper/internal/api"
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
	// ItemCreate creates a new item in the vault
	ItemCreate(title string, itemType models.ItemType, fields ...*models.Field) (*models.ItemOverview, error)
	// ItemUpdate updates an existing item in the vault
	ItemUpdate(id []byte, title string, fields ...*models.Field) (*models.ItemOverview, error)
	// ItemDelete deletes an item from the vault
	ItemDelete(id []byte) error
	// ItemGet returns an item from the vault
	ItemGet(id []byte) (*models.Item, error)
	// ItemOverviewList returns a list of overviews of all items in the vault.
	ItemOverviewList() ([]models.ItemOverview, error)
	// ItemOverviewFilter returns a list of overviews of all items in the vault that match the filter
	ItemOverviewFilter(filter string) ([]models.ItemOverview, error)
	// ItemAttachmentAdd adds an attachment to an item
	ItemAttachmentAdd(id, attachmentID []byte) error
	// ItemAttachmentRemove removes an attachment from an item
	ItemAttachmentRemove(id, attachmentID []byte) error
	// ItemSyncOut returns a list of items that have been changed since the given timestamp
	ItemSyncOut(since int64) ([]api.SyncItem, error)
	// ItemSyncIn updates the vault with the given sync item
	ItemSyncIn(syncItem *api.SyncItem) error
}

// VaultAttachments is the interface for managing attachments in the vault
type VaultAttachments interface {
	// AttachmentAdd adds a new attachment to the vault
	AttachmentAdd(fileName string, r io.Reader) (*models.AttachmentOverview, error)
	// AttachmentRemove marks an attachment as deleted
	AttachmentRemove(id []byte) error
	// AttachmentOverviewGet returns an overview of an attachment
	AttachmentOverviewGet(id []byte) (*models.AttachmentOverview, error)
	// AttachmentDataGet returns the data of an attachment
	AttachmentDataGet(id []byte) (io.Reader, error)
	// AttachmentSyncOut returns a list of encrypted attachments that have been changed since the given timestamp
	AttachmentSyncOut(since int64) ([]api.SyncAttachment, error)
	// AttachmentSyncIn updates the vault with the given sync attachment
	AttachmentSyncIn(syncAttachment *api.SyncAttachment) error
}

// VaultSyncServer is the interface for managing the sync server configuration
type VaultSyncServer interface {
	// SyncServerGet returns the sync server configuration
	SyncServerGet() (*models.SyncServer, error)
	// SyncServerSet sets the sync server configuration
	SyncServerSet(srv *models.SyncServer) error
	// SyncServerRefreshTokenSet sets the refresh token for the sync server
	SyncServerRefreshTokenSet(token []byte) error
	// SyncServerLastSyncSet sets the timestamp of the last sync
	SyncServerLastSyncSet(timestamp int64) error
}

// VaultKeys is the interface for managing the vault keys
type VaultKeys interface {
	// GenerateNewKeys generates new keys for the vault and encrypts them with the given master password
	GenerateNewKeys(masterPass []byte) error
	// ReEncryptKeys re-encrypts the keys with the new master password
	ReEncryptKeys(oldMasterPass, newMasterPass []byte) error
	// HasKeys returns true if the vault keys are available
	HasKeys() bool
	// KeysSyncOut returns the encrypted keys if they have been changed since the given timestamp
	KeysSyncOut(since int64) (*api.SyncKeys, error)
	// KeysSyncIn updates the vault with the given encrypted keys
	KeysSyncIn(newMasterPass []byte, newKeys *api.SyncKeys) error
}
