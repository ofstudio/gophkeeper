package boltvault

import (
	"google.golang.org/protobuf/proto"

	"github.com/ofstudio/gophkeeper/internal/client/repo"
)

func (suite *boltVaultSuite) TestSyncServerGet() {
	suite.Run("empty repo", func() {
		r := suite.initRepoWithKeys("password")
		_, err := r.SyncServerGet()
		suite.ErrorIs(err, repo.ErrNotFound)
	})

	suite.Run("successful get", func() {
		r := suite.initRepoWithKeys("password")
		suite.Require().NoError(r.SyncServerPut(syncServer1))
		s, err := r.SyncServerGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(s)
		suite.True(proto.Equal(syncServer1, s))
	})
}

func (suite *boltVaultSuite) TestSyncServerPut() {
	suite.Run("successful put", func() {
		r := suite.initRepoWithKeys("password")
		suite.Require().NoError(r.SyncServerPut(syncServer1))

		s, err := r.SyncServerGet()
		suite.Require().NoError(err)
		suite.Require().NotNil(s)
		suite.True(proto.Equal(syncServer1, s))
	})
}

func (suite *boltVaultSuite) TestSyncServerPurge() {
	suite.Run("successful purge", func() {
		r := suite.initRepoWithKeys("password")
		suite.Require().NoError(r.SyncServerPut(syncServer1))
		suite.Require().NoError(r.SyncServerPurge())

		_, err := r.SyncServerGet()
		suite.ErrorIs(err, repo.ErrNotFound)
	})

	suite.Run("purge non-existing sync server", func() {
		r := suite.initRepoWithKeys("password")
		err := r.SyncServerPurge()
		suite.Require().NoError(err)
	})
}