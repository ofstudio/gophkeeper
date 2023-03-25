package boltvault

import (
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/providers/encrypt"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

func (suite *boltVaultSuite) TestAttachmentAdd() {
	suite.Run("successful add", func() {
		rt := suite.timeNow(10)
		defer rt()
		r := suite.initRepoWithKeys("password")

		m, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)
		suite.Require().NotNil(m)

		m, err = r.AttachmentMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(m)
		suite.True(proto.Equal(file1Meta, m))

		d, err := r.AttachmentDataGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(d)
		suite.Equal(file1Content, string(d))
	})

}

func (suite *boltVaultSuite) TestAttachmentDelete() {
	suite.Run("successful remove", func() {
		r := suite.initRepoWithKeys("password")

		m, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		err = r.AttachmentDelete(m.Id)
		suite.Require().NoError(err)

		md, err := r.AttachmentMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.Equal(true, md.Deleted)
		suite.Equal("", md.FileName)

		d, err := r.AttachmentDataGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(d)
		suite.Require().NoError(err)
		suite.Len(d, 0)
	})

	suite.Run("remove non-existing attachment", func() {
		r := suite.initRepoWithKeys("password")
		err := r.AttachmentDelete([]byte("non-existing-id"))
		suite.Require().Equal(repo.ErrNotFound, err)
	})
}

func (suite *boltVaultSuite) TestAttachmentMetaGet() {
	suite.Run("successful get", func() {
		r := suite.initRepoWithKeys("password")

		m, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		mg, err := r.AttachmentMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.True(proto.Equal(m, mg))
	})

	suite.Run("get non-existing attachment", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.AttachmentMetaGet([]byte("non-existing-id"))
		suite.Require().Equal(repo.ErrNotFound, err)
	})
}

func (suite *boltVaultSuite) TestAttachmentDataGet() {
	suite.Run("successful get", func() {
		r := suite.initRepoWithKeys("password")

		o, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		d, err := r.AttachmentDataGet(o.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(d)
		suite.Equal(file1Content, string(d))
	})

	suite.Run("get non-existing attachment", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.AttachmentDataGet([]byte("non-existing-id"))
		suite.Require().Equal(repo.ErrNotFound, err)
	})
}

func (suite *boltVaultSuite) TestAttachmentEncryption() {
	suite.Run("should encrypt when using AESGCM", func() {
		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewAESGCM())
		suite.Require().NoError(r.KeysGenerateNew([]byte("password")))
		suite.Require().NoError(r.Unlock([]byte("password")))
		m, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		// Check that the data is encrypted
		var bm, bd []byte
		_ = r.db.View(func(tx *bbolt.Tx) error {
			bm, err = r.txValueGet(tx, bucketAttachmentsMeta, m.Id)
			suite.Require().NoError(err)
			bd, err = r.txValueGet(tx, bucketAttachmentsData, m.Id)
			suite.Require().NoError(err)
			return nil
		})
		suite.Error(proto.Unmarshal(bm, &models.AttachmentMeta{}))
		suite.NotEqual(file1Content, string(bd))
	})

	suite.Run("should not encrypt when using Nop", func() {
		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewNop())
		suite.Require().NoError(r.KeysGenerateNew([]byte("password")))
		suite.Require().NoError(r.Unlock([]byte("password")))
		m, err := r.AttachmentPut(file1Meta, []byte(file1Content))
		suite.Require().NoError(err)

		// Check that the data is encrypted
		var bm, bd []byte
		_ = r.db.View(func(tx *bbolt.Tx) error {
			bm, err = r.txValueGet(tx, bucketAttachmentsMeta, m.Id)
			suite.Require().NoError(err)
			bd, err = r.txValueGet(tx, bucketAttachmentsData, m.Id)
			suite.Require().NoError(err)
			return nil
		})
		suite.NoError(proto.Unmarshal(bm, &models.AttachmentMeta{}))
		suite.Equal(file1Content, string(bd))
	})
}

func (suite *boltVaultSuite) TestAttachmentVacuum() {
	suite.Run("no attachments", func() {
		r := suite.initRepoWithKeys("password")
		suite.NoError(r.attachmentVacuum())
	})
	suite.Run("no attachments to delete", func() {
		r := suite.initRepoWithKeys("password")
		ma1, _ := r.AttachmentPut(file1Meta, []byte(file1Content))
		ma2, _ := r.AttachmentPut(file2Meta, []byte(file2Content))

		suite.NoError(r.attachmentVacuum())
		_, err := r.AttachmentMetaGet(ma1.Id)
		suite.NoError(err)
		_, err = r.AttachmentMetaGet(ma2.Id)
		suite.NoError(err)
	})

	suite.Run("attachments marked as deleted", func() {
		r := suite.initRepoWithKeys("password")
		ma1, _ := r.AttachmentPut(file1Meta, []byte(file1Content))
		ma2, _ := r.AttachmentPut(file2Meta, []byte(file2Content))
		suite.Require().NoError(r.AttachmentDelete(ma1.Id))
		suite.Require().NoError(r.AttachmentDelete(ma2.Id))

		suite.NoError(r.attachmentVacuum())
		_, err := r.AttachmentMetaGet(ma1.Id)
		suite.ErrorIs(err, repo.ErrNotFound)
		_, err = r.AttachmentMetaGet(ma2.Id)
		suite.ErrorIs(err, repo.ErrNotFound)
	})

}
