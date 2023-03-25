package boltvault

import (
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"github.com/ofstudio/gophkeeper/internal/client/models"
	"github.com/ofstudio/gophkeeper/internal/client/providers/encrypt"
	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

func (suite *boltVaultSuite) TestItemPut() {
	suite.Run("successful create", func() {
		rt := suite.timeNow(12345)
		defer rt()
		r := suite.initRepoWithKeys("master password")
		suite.Require().NoError(r.Unlock([]byte("master password")))

		m, err := r.ItemPut(item1Meta, item1Data)
		suite.Require().NoError(err)
		suite.NotNil(m)
		suite.NotEmpty(m.Id)

		m, err = r.ItemMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.NotNil(m)
		suite.True(proto.Equal(item1Meta, m))

		d, err := r.ItemDataGet(m.Id)
		suite.Require().NoError(err)
		suite.NotNil(d)
		suite.True(proto.Equal(item1Data, d))
	})

	suite.Run("successful update", func() {
		r := suite.initRepoWithKeys("master password")
		m, err := r.ItemPut(item1Meta, item1Data)
		suite.Require().NoError(err)
		suite.NotNil(m)

		m.Title = "Updated title"
		m, err = r.ItemPut(m, item2Data)
		suite.Require().NoError(err)

		m, err = r.ItemMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.NotNil(m)
		suite.Equal("Updated title", m.Title)

		d, err := r.ItemDataGet(m.Id)
		suite.Require().NoError(err)
		suite.NotNil(d)
		suite.True(proto.Equal(item2Data, d))
	})

	suite.Run("nil arguments", func() {
		r := suite.initRepoWithKeys("master password")
		_, err := r.ItemPut(nil, &models.ItemData{})
		suite.ErrorIs(err, repo.ErrInvalidArgument)
		_, err = r.ItemPut(&models.ItemMeta{}, nil)
		suite.ErrorIs(err, repo.ErrInvalidArgument)
		_, err = r.ItemPut(nil, nil)
		suite.ErrorIs(err, repo.ErrInvalidArgument)
	})
}

func (suite *boltVaultSuite) TestItemDelete() {
	suite.Run("successful delete", func() {
		r := suite.initRepoWithKeys("password")
		rt := suite.timeNow(12345)
		defer rt()
		m, _ := r.ItemPut(item1Meta, item1Data)

		suite.timeNow(23456)
		suite.Require().NoError(r.ItemDelete(m.Id))

		md, err := r.ItemMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(md)
		suite.True(proto.Equal(md, &models.ItemMeta{
			Id:          m.Id,
			CreatedAt:   0,
			UpdatedAt:   23456,
			Title:       "",
			Type:        0,
			Deleted:     true,
			Attachments: nil,
		}))

		dd, err := r.ItemDataGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(dd)
		suite.Nil(dd.Fields)
	})

	suite.Run("non existing item", func() {
		r := suite.initRepoWithKeys("password")
		suite.ErrorIs(r.ItemDelete([]byte("non existing item")), repo.ErrNotFound)
	})
}

func (suite *boltVaultSuite) TestItemMetaGet() {
	suite.Run("successful get", func() {
		r := suite.initRepoWithKeys("password")
		rt := suite.timeNow(12345)
		defer rt()
		m, _ := r.ItemPut(item2Meta, item2Data)

		mg, err := r.ItemMetaGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(mg)
		suite.Equal(m.Id, mg.Id)
	})

	suite.Run("non existing item", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.ItemMetaGet([]byte("non existing item"))
		suite.ErrorIs(err, repo.ErrNotFound)
	})
}

func (suite *boltVaultSuite) TestItemDataGet() {
	suite.Run("successful get", func() {
		r := suite.initRepoWithKeys("password")
		rt := suite.timeNow(12345)
		defer rt()
		m, _ := r.ItemPut(item2Meta, item2Data)

		dg, err := r.ItemDataGet(m.Id)
		suite.Require().NoError(err)
		suite.Require().NotNil(dg)
		suite.True(proto.Equal(item2Data, dg))
	})

	suite.Run("non existing item", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.ItemDataGet([]byte("non existing item"))
		suite.ErrorIs(err, repo.ErrNotFound)
	})
}

func (suite *boltVaultSuite) TestItemMetaList() {
	r := suite.initRepoWithKeys("password")
	m1, _ := r.ItemPut(item2Meta, item2Data)
	m2, _ := r.ItemPut(item1Meta, item1Data)

	l, err := r.ItemMetaList()
	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 2)

	suite.Require().NoError(r.ItemDelete(m1.Id))
	l, err = r.ItemMetaList()

	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 1)

	suite.Equal(m2.Id, l[0].Id)
	suite.Equal("Login1", l[0].Title)
}

func (suite *boltVaultSuite) TestItemMetaFilter() {
	r := suite.initRepoWithKeys("password")
	_, _ = r.ItemPut(item2Meta, item2Data)
	_, _ = r.ItemPut(item1Meta, item1Data)

	l, err := r.ItemMetaFilter("")
	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 2)

	l, err = r.ItemMetaFilter("1")
	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 2)

	l, err = r.ItemMetaFilter("Log")
	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 1)
	suite.Equal("Login1", l[0].Title)

	l, err = r.ItemMetaFilter("Not")
	suite.Require().NoError(err)
	suite.Require().NotNil(l)
	suite.Len(l, 1)
	suite.Equal("Note1", l[0].Title)

	l, err = r.ItemMetaFilter("Not exist")
	suite.Require().NoError(err)
	suite.Nil(l)
}

func (suite *boltVaultSuite) TestItemEncryption() {

	suite.Run("should encrypt when using AESGCM", func() {
		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewAESGCM())
		suite.Require().NoError(r.KeysGenerateNew([]byte("password")))
		suite.Require().NoError(r.Unlock([]byte("password")))
		m, err := r.ItemPut(item2Meta, item2Data)
		suite.Require().NoError(err)

		// Check that the data is encrypted
		var bm, bd []byte
		_ = r.db.View(func(tx *bbolt.Tx) error {
			bm, err = r.txValueGet(tx, bucketItemsMeta, m.Id)
			suite.Require().NoError(err)
			bd, err = r.txValueGet(tx, bucketItemsData, m.Id)
			suite.Require().NoError(err)
			return nil
		})
		suite.Error(proto.Unmarshal(bm, &models.ItemMeta{}))
		suite.Error(proto.Unmarshal(bd, &models.ItemData{}))
	})

	suite.Run("should not encrypt when using Nop", func() {
		r := suite.initRepo()
		r.WithEncryptor(encrypt.NewNop())
		suite.Require().NoError(r.KeysGenerateNew([]byte("password")))
		suite.Require().NoError(r.Unlock([]byte("password")))
		m, err := r.ItemPut(item2Meta, item2Data)
		suite.Require().NoError(err)
		// Check that the data is encrypted
		var bm, bd []byte
		_ = r.db.View(func(tx *bbolt.Tx) error {
			bm, err = r.txValueGet(tx, bucketItemsMeta, m.Id)
			suite.Require().NoError(err)
			bd, err = r.txValueGet(tx, bucketItemsData, m.Id)
			suite.Require().NoError(err)
			return nil
		})
		suite.NoError(proto.Unmarshal(bm, &models.ItemMeta{}))
		suite.NoError(proto.Unmarshal(bd, &models.ItemData{}))
	})
}
