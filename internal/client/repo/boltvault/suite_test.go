package boltvault

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.etcd.io/bbolt"

	"github.com/ofstudio/gophkeeper/internal/client/providers/encrypt"
)

func TestSuite(t *testing.T) {
	suite.Run(t, new(boltVaultSuite))
}

type boltVaultSuite struct {
	suite.Suite
	tempDir string
	repos   []*BoltVault
}

func (suite *boltVaultSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "boltvault_test-*")
	suite.Require().NoError(err)
	suite.repos = make([]*BoltVault, 0)
}

func (suite *boltVaultSuite) TearDownTest() {
	for _, r := range suite.repos {
		suite.Require().NoError(r.Close())
	}
	suite.Require().NoError(os.RemoveAll(suite.tempDir))
}

// initRepo initializes the repository for testing without generating keys.
func (suite *boltVaultSuite) initRepo() *BoltVault {
	r, err := NewBoltVault(path.Join(suite.tempDir, suite.randFName("test.db")))
	suite.Require().NoError(err)
	r.WithEncryptor(encrypt.NewAESGCM())

	suite.repos = append(suite.repos, r)
	return r
}

// initRepoWithKeys initializes the repository for testing,
// generates keys with the given master password and unlocks the repository.
func (suite *boltVaultSuite) initRepoWithKeys(masterPass string) *BoltVault {
	r := suite.initRepo()
	mp := []byte(masterPass)
	suite.Require().NoError(r.KeysGenerateNew(mp))
	suite.Require().NoError(r.Unlock(mp))
	return r
}

// randFName generates a random filename with the given suffix
func (suite *boltVaultSuite) randFName(suffix string) string {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	suite.Require().NoError(err)
	return fmt.Sprintf("%x_%s", randBytes, suffix)
}

// isWiped checks if the given byte slice is wiped (all bytes are 0)
func (suite *boltVaultSuite) isWiped(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			suite.T().Logf("%v is not wiped", b)
			suite.T().Fail()
			return false
		}
	}
	return true
}

// isNotWiped checks if the given byte slice is not wiped (at least one byte is not 0)
func (suite *boltVaultSuite) isNotWiped(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return true
		}
	}
	suite.T().Logf("%v is wiped", b)
	suite.T().Fail()
	return false
}

// isEmptyRepo checks if the given repository is empty
func (suite *boltVaultSuite) isEmptyRepo(r *BoltVault) error {
	return r.db.View(func(tx *bbolt.Tx) error {
		for _, bn := range vaultBuckets {
			b := tx.Bucket(bn)
			if b == nil {
				return fmt.Errorf("bucket %s not found", bn)
			}
			return b.ForEach(func(k, _ []byte) error {
				if !bytes.Equal(bn, bucketSettings) && !bytes.Equal(k, keyDBVersion) {
					return fmt.Errorf("bucket %s contains key %s", bn, k)
				}
				return nil
			})
		}
		return nil
	})
}

// timeNow sets the nowFunc to the given Unix timestamp
// and returns a function to restore the original nowFunc
func (suite *boltVaultSuite) timeNow(ts int64) func() {
	origNowFunc := nowFunc
	nowFunc = func() time.Time {
		return time.Unix(ts, 0)
	}
	return func() {
		nowFunc = origNowFunc
	}
}
