package boltvault

import (
	"path"

	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

func (suite *boltVaultSuite) TestNewBoltVault() {
	suite.Run("invalid file path", func() {
		_, err := NewBoltVault("")
		suite.ErrorIs(err, repo.ErrDBFailedToOpen)

		_, err = NewBoltVault("*/**/invalid")
		suite.ErrorIs(err, repo.ErrDBFailedToOpen)
	})

	suite.Run("migration from unsupported version", func() {
		// Create a database with unsupported version
		dbFilePath := path.Join(suite.tempDir, "/unsupported.db")
		db, err := bbolt.Open(dbFilePath, 0600, nil)
		suite.Require().NoError(err)
		_ = db.Update(func(tx *bbolt.Tx) error {
			_, err = tx.CreateBucketIfNotExists(bucketSettings)
			suite.Require().NoError(err)
			err = tx.Bucket(bucketSettings).Put(keyDBVersion, []byte("0.0.0"))
			suite.Require().NoError(err)
			return nil
		})
		suite.Require().NoError(db.Close())

		// Try to open database
		_, err = NewBoltVault(dbFilePath)
		suite.ErrorIs(err, repo.ErrDBVersionNotSupported)
	})
}

func (suite *boltVaultSuite) TestIsLocked() {
	suite.Run("locked repo", func() {
		r := suite.initRepo()
		suite.True(r.IsLocked())

		suite.NoError(r.KeysGenerateNew([]byte("password")))
		suite.True(r.IsLocked())
	})

	suite.Run("unlocked repo", func() {
		r := suite.initRepoWithKeys("password")

		suite.NoError(r.Unlock([]byte("password")))
		suite.False(r.IsLocked())
	})
}

func (suite *boltVaultSuite) TestLock() {
	suite.Run("locked repo", func() {
		r := suite.initRepo()
		suite.NoError(r.KeysGenerateNew([]byte("password")))

		suite.True(r.IsLocked())
		r.Lock()
		suite.True(r.IsLocked())
		suite.Nil(r.masterKeyEnc)
		suite.Nil(r.salt)
	})

	suite.Run("unlocked repo", func() {
		r := suite.initRepoWithKeys("password")

		suite.NoError(r.Unlock([]byte("password")))
		suite.False(r.IsLocked())
		salt := r.salt
		suite.isNotWiped(salt)

		r.Lock()
		suite.True(r.IsLocked())
		suite.Nil(r.masterKeyEnc)
		suite.Nil(r.salt)
		suite.isWiped(salt)
	})
}

func (suite *boltVaultSuite) TestUnlock() {
	suite.Run("empty repo", func() {
		r := suite.initRepo()
		suite.True(r.IsLocked())
		err := r.Unlock([]byte("password"))
		suite.ErrorIs(err, repo.ErrNotFound)
		suite.True(r.IsLocked())
	})

	suite.Run("successful unlock", func() {
		r := suite.initRepo()
		suite.NoError(r.KeysGenerateNew([]byte("password")))
		suite.True(r.IsLocked())
		suite.NoError(r.Unlock([]byte("password")))
		suite.False(r.IsLocked())
	})

	suite.Run("wrong master password", func() {
		r := suite.initRepoWithKeys("password")
		r.Lock()

		suite.True(r.IsLocked())
		err := r.Unlock([]byte("wrong password"))
		suite.ErrorIs(err, repo.ErrFailedToDecrypt)
		err = r.Unlock([]byte{})
		suite.ErrorIs(err, repo.ErrFailedToDecrypt)
		err = r.Unlock(nil)
		suite.ErrorIs(err, repo.ErrFailedToDecrypt)
	})

}

func (suite *boltVaultSuite) TestVacuum() {
	suite.Run("empty repo", func() {
		r := suite.initRepo()
		suite.NoError(r.Vacuum())
		r = suite.initRepoWithKeys("password")
		suite.NoError(r.Vacuum())
	})

	suite.Run("nothing to vacuum", func() {
		r := suite.initRepoWithKeys("password")
		i, err := r.ItemPut(item1Meta, item1Data)
		suite.Require().NoError(err)
		a, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		suite.NoError(r.Vacuum())
		suite.NoError(r.Unlock([]byte("password")))
		_, err = r.ItemMetaGet(i.Id)
		suite.NoError(err)
		_, err = r.ItemDataGet(i.Id)
		suite.NoError(err)
		_, err = r.AttachmentMetaGet(a.Id)
		suite.NoError(err)
		_, err = r.AttachmentDataGet(a.Id)
		suite.NoError(err)
	})

	suite.Run("vacuum some items", func() {
		r := suite.initRepoWithKeys("password")
		i1, err := r.ItemPut(item1Meta, item1Data)
		suite.Require().NoError(err)
		i2, err := r.ItemPut(item2Meta, item2Data)
		suite.Require().NoError(err)
		a1, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)
		a2, err := r.AttachmentPut(file2Meta, []byte(file2Content))
		suite.Require().NoError(err)

		suite.NoError(r.ItemDelete(i1.Id))
		suite.NoError(r.AttachmentDelete(a1.Id))

		suite.NoError(r.Vacuum())
		suite.NoError(r.Unlock([]byte("password")))

		_, err = r.ItemMetaGet(i1.Id)
		suite.ErrorIs(err, repo.ErrNotFound)
		_, err = r.ItemDataGet(i1.Id)
		suite.ErrorIs(err, repo.ErrNotFound)
		_, err = r.AttachmentMetaGet(a1.Id)
		suite.ErrorIs(err, repo.ErrNotFound)
		_, err = r.AttachmentDataGet(a1.Id)
		suite.ErrorIs(err, repo.ErrNotFound)

		_, err = r.ItemMetaGet(i2.Id)
		suite.NoError(err)
		_, err = r.ItemDataGet(i2.Id)
		suite.NoError(err)
		_, err = r.AttachmentMetaGet(a2.Id)
		suite.NoError(err)
		_, err = r.AttachmentDataGet(a2.Id)
		suite.NoError(err)
	})
}

func (suite *boltVaultSuite) TestPurge() {
	suite.Run("empty repo", func() {
		r := suite.initRepoWithKeys("password")
		suite.NoError(r.Purge())
		suite.True(r.IsLocked())
		suite.NoError(suite.isEmptyRepo(r))
	})

	suite.Run("purge repo with items ", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.ItemPut(item1Meta, item1Data)
		suite.Require().NoError(err)
		_, err = r.ItemPut(item2Meta, item2Data)
		suite.Require().NoError(err)
		_, err = r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)
		_, err = r.AttachmentPut(file2Meta, []byte(file2Content))

		suite.NoError(r.Purge())
		suite.True(r.IsLocked())
		suite.NoError(suite.isEmptyRepo(r))
	})
}
