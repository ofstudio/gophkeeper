package boltvault

import "github.com/ofstudio/gophkeeper/internal/client/models"

var (
	item1Meta, item2Meta *models.ItemMeta
	item1Data, item2Data *models.ItemData
	file1Meta, file2Meta *models.AttachmentMeta
	syncServer1          *models.SyncServer
)

const (
	file1Content = "This is a test file. 🙂"
	file2Content = "This is another test file. 😇"
)

func (suite *boltVaultSuite) SetupSubTest() {
	item1Meta = &models.ItemMeta{Title: "Login1", Type: models.ItemType_LOGIN}
	item1Data = &models.ItemData{
		Fields: []*models.Field{
			{Order: 0, Title: "login", Type: models.FieldType_TEXT, Value: []byte("username")},
			{Order: 1, Title: "password", Type: models.FieldType_SECRET, Value: []byte("password")},
			{Order: 2, Title: "url", Type: models.FieldType_URL, Value: []byte("https://example.com")},
		},
	}

	item2Meta = &models.ItemMeta{Title: "Note1", Type: models.ItemType_SECURE_NOTE}
	item2Data = &models.ItemData{
		Fields: []*models.Field{
			{Order: 0, Title: "note", Type: models.FieldType_NOTE, Value: []byte("Lorem ipsum dolor sit amet")},
		},
	}

	file1Meta = &models.AttachmentMeta{FileName: "file1.txt", FileSize: uint64(len(file1Content))}
	file2Meta = &models.AttachmentMeta{FileName: "file2.txt", FileSize: uint64(len(file2Content))}

	syncServer1 = &models.SyncServer{
		Url:          "http://localhost:8080",
		Username:     "username",
		RefreshToken: []byte("refresh-token-data"),
		LastSyncedAt: 12345,
	}

}
