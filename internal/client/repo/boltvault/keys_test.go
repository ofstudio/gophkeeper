package boltvault

import (
	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/providers/encrypt"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

func (suite *boltVaultSuite) TestKeysGenerateNew() {
	suite.Run("success", func() {
		rt := suite.timeNow(1234)
		defer rt()

		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewAESGCM())
		suite.Require().Nil(r.masterKeyEnc)
		suite.Require().Nil(r.salt)

		// Generate new keys
		suite.Require().NoError(r.KeysGenerateNew([]byte("test")))

		_ = r.db.View(func(tx *bbolt.Tx) error {
			// Retrieve keys from vault
			keys, err := r.txVaultKeysGet(tx)
			suite.Require().NoError(err)
			suite.Require().NotNil(keys)
			suite.Require().NotNil(keys.MasterKeyEncrypted)
			suite.Require().NotNil(keys.Salt)
			suite.Require().Equal(int64(1234), keys.UpdatedAt)

			// Check master key
			mk, err := r.encryptor.DecryptMasterKey(keys.MasterKeyEncrypted, []byte("test"), keys.Salt)
			suite.Require().NoError(err)
			suite.Require().NotNil(mk)
			suite.Require().NotEqual(keys.MasterKeyEncrypted, mk)

			return nil
		})

	})

	suite.Run("already exists", func() {
		r := suite.initRepo()
		suite.Require().NoError(r.KeysGenerateNew([]byte("test1")))
		suite.Require().ErrorIs(r.KeysGenerateNew([]byte("test2")), repo.ErrAlreadyExists)
	})

	suite.Run("missing encryptor", func() {
		r := suite.initRepo()
		r.encryptor = nil
		suite.Require().ErrorIs(r.KeysGenerateNew([]byte("test")), repo.ErrMissingEncryptProvider)
	})

	suite.Run("zero length master password", func() {
		r := suite.initRepo()
		suite.Require().ErrorIs(r.KeysGenerateNew([]byte{}), repo.ErrFailedToEncrypt)
		r = suite.initRepo()
		suite.Require().ErrorIs(r.KeysGenerateNew(nil), repo.ErrFailedToEncrypt)
	})
}

func (suite *boltVaultSuite) TestKeysPut() {
	suite.Run("empty repo", func() {
		// Create repo with keys
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)

		// Create new empty repo
		r = suite.initRepo()
		suite.Require().False(r.KeysExist())
		suite.Require().NoError(r.KeysPut([]byte("password"), k))

		// Check keys
		kk, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(kk)
		suite.Require().Equal(k.MasterKeyEncrypted, kk.MasterKeyEncrypted)
	})

	suite.Run("with existing keys", func() {
		// Create repo with keys
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)

		// Re-encrypt keys
		suite.Require().NoError(r.KeysReEncrypt([]byte("password"), []byte("password2")))
		suite.Require().NoError(r.Unlock([]byte("password2")))
		k2, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k2)
		suite.Require().NotEqual(k.MasterKeyEncrypted, k2.MasterKeyEncrypted)

		// Put keys
		suite.Require().NoError(r.KeysPut([]byte("password"), k))
		k3, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k3)
		suite.Require().Equal(k.MasterKeyEncrypted, k3.MasterKeyEncrypted)
	})

	suite.Run("wrong master password", func() {
		// Create repo with keys
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)

		// Put keys
		suite.Require().ErrorIs(r.KeysPut([]byte("password2"), k), repo.ErrFailedToDecrypt)
	})

	suite.Run("master key mismatch", func() {
		// Create repo with keys
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)

		// Put keys
		k.MasterKeyEncrypted = []byte("wrong master key encrypted is here :)")
		suite.Require().ErrorIs(r.KeysPut([]byte("password"), k), repo.ErrFailedToDecrypt)
	})

	suite.Run("missing encryptor", func() {
		// Create repo with keys
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)

		// Put keys
		r.encryptor = nil
		suite.Require().ErrorIs(r.KeysPut([]byte("password"), k), repo.ErrMissingEncryptProvider)
	})
}

func (suite *boltVaultSuite) TestKeysGet() {
	suite.Run("success", func() {
		r := suite.initRepoWithKeys("password")
		k, err := r.KeysGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(k)
		suite.NotNil(k.MasterKeyEncrypted)
		suite.NotNil(k.Salt)
		suite.NotNil(k.UpdatedAt)
	})

	suite.Run("missing keys", func() {
		r := suite.initRepo()
		_, err := r.KeysGet()
		suite.ErrorIs(err, repo.ErrNotFound)
	})
}

func (suite *boltVaultSuite) TestKeysReEncrypt() {
	suite.Run("success", func() {
		r := suite.initRepo()
		var keys1, keys2 *models.Keys
		var mk1, mk2 []byte
		var err error

		// Generate new keys
		suite.Require().NoError(r.KeysGenerateNew([]byte("test1")))

		_ = r.db.View(func(tx *bbolt.Tx) error {
			keys1, err = r.txVaultKeysGet(tx)
			suite.Require().NoError(err)
			suite.Require().NotNil(keys1)
			return nil
		})
		// Decrypt master key
		mk1, err = r.encryptor.DecryptMasterKey(keys1.MasterKeyEncrypted, []byte("test1"), keys1.Salt)
		suite.Require().NoError(err)

		// Re-encrypt keys
		suite.Require().NoError(r.KeysReEncrypt([]byte("test1"), []byte("test2")))

		_ = r.db.View(func(tx *bbolt.Tx) error {
			// Retrieve keys from vault
			keys2, err = r.txVaultKeysGet(tx)
			suite.Require().NoError(err)
			suite.Require().NotNil(keys2)
			return nil
		})

		// Decrypt master key
		mk2, err = r.encryptor.DecryptMasterKey(keys2.MasterKeyEncrypted, []byte("test2"), keys2.Salt)
		suite.Require().NoError(err)

		// Compare keys
		suite.Require().Equal(mk1, mk2)
		suite.Require().NotEqual(keys1.MasterKeyEncrypted, keys2.MasterKeyEncrypted)
		suite.Require().NotEqual(keys1.Salt, keys2.Salt)
	})

	suite.Run("no keys", func() {
		r := suite.initRepo()
		suite.Require().ErrorIs(r.KeysReEncrypt([]byte("test1"), []byte("test2")), repo.ErrNotFound)
	})
}

func (suite *boltVaultSuite) TestKeysExist() {
	suite.Run("success", func() {
		r := suite.initRepo()
		suite.Require().False(r.KeysExist())
		suite.Require().NoError(r.KeysGenerateNew([]byte("test")))
		suite.Require().True(r.KeysExist())
	})
}

func (suite *boltVaultSuite) TestKeysEncryption() {
	suite.Run("should encrypt when using aes-gcm", func() {
		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewAESGCM())
		suite.Require().NoError(r.KeysGenerateNew([]byte("test")))
		suite.Require().NoError(r.Unlock([]byte("test")))
		mk, err := r.masterKeyEnc.Open()
		suite.Require().NoError(err)
		suite.Require().NotNil(mk)

		_ = r.db.View(func(tx *bbolt.Tx) error {
			keys, err := r.txVaultKeysGet(tx)
			suite.Require().NoError(err)
			suite.Require().NotNil(keys)
			suite.Require().NotEqual(mk.Bytes(), keys.MasterKeyEncrypted)
			return nil
		})
	})

}
