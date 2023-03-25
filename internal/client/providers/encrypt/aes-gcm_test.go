package encrypt

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSuite(t *testing.T) {
	suite.Run(t, new(aesGCMSuite))
}

type aesGCMSuite struct {
	suite.Suite
}

func (suite *aesGCMSuite) TestNewKey() {
	suite.Run("successful random key generation", func() {
		p := NewAESGCM()
		k1, err := p.NewKey()
		suite.NoError(err)
		suite.Len(k1, 32)
		k2, err := p.NewKey()
		suite.NoError(err)
		suite.Len(k2, 32)
		suite.NotEqual(k1, k2)
	})
}

func (suite *aesGCMSuite) TestNewSalt() {
	suite.Run("successful random salt generation", func() {
		p := NewAESGCM()
		s1, err := p.NewSalt()
		suite.NoError(err)
		suite.Len(s1, 32)
		s2, err := p.NewSalt()
		suite.NoError(err)
		suite.Len(s2, 32)
		suite.NotEqual(s1, s2)
	})
}

func (suite *aesGCMSuite) TestEncryptData() {
	suite.Run("successful encryption", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		d := []byte("test data")
		ed, err := p.EncryptData(d, k)
		suite.NoError(err)
		suite.NotEqual(d, ed)
	})

	suite.Run("wrong key length", func() {
		p := NewAESGCM()
		k := []byte("test key")
		d := []byte("test data")
		ed, err := p.EncryptData(d, k)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(ed)
	})

	suite.Run("nil key", func() {
		p := NewAESGCM()
		d := []byte("test data")
		ed, err := p.EncryptData(d, nil)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(ed)
	})

	suite.Run("nil data", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		ed, err := p.EncryptData(nil, k)
		suite.ErrorIs(err, ErrNoData)
		suite.Nil(ed)
	})

	suite.Run("empty data", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		d := make([]byte, 0)
		ed, err := p.EncryptData(d, k)
		suite.NoError(err)
		suite.NotNil(ed)
		suite.NotEqual(d, ed)
	})
}

func (suite *aesGCMSuite) TestDecryptData() {
	suite.Run("successful decryption", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		d := []byte("test data")
		ed, err := p.EncryptData(d, k)
		suite.NoError(err)
		suite.NotEqual(d, ed)
		dd, err := p.DecryptData(ed, k)
		suite.NoError(err)
		suite.Equal(d, dd)
	})

	suite.Run("wrong key length", func() {
		p := NewAESGCM()
		k := []byte("test key")
		d := []byte("test data")
		dd, err := p.DecryptData(d, k)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(dd)
	})

	suite.Run("nil key", func() {
		p := NewAESGCM()
		d := []byte("test data")
		dd, err := p.DecryptData(d, nil)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(dd)
	})

	suite.Run("nil data", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		dd, err := p.DecryptData(nil, k)
		suite.ErrorIs(err, ErrNoData)
		suite.Nil(dd)
	})

	suite.Run("empty data", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		d := make([]byte, 0)
		ed, err := p.EncryptData(d, k)
		suite.NoError(err)
		suite.NotNil(ed)
		suite.NotEqual(d, ed)
		dd, err := p.DecryptData(ed, k)
		suite.NoError(err)
		suite.Equal(d, dd)
	})

	suite.Run("wrong key", func() {
		p := NewAESGCM()
		k1, err := p.NewKey()
		suite.NoError(err)
		k2, err := p.NewKey()
		suite.NoError(err)
		d := []byte("test data")
		ed, err := p.EncryptData(d, k1)
		suite.NoError(err)
		suite.NotEqual(d, ed)
		dd, err := p.DecryptData(ed, k2)
		suite.ErrorIs(err, ErrFailedToDecrypt)
		suite.Nil(dd)
	})

	suite.Run("wrong data", func() {
		p := NewAESGCM()
		k, err := p.NewKey()
		suite.NoError(err)
		d := []byte("test data")
		ed, err := p.EncryptData(d, k)
		suite.NoError(err)
		suite.NotEqual(d, ed)
		ed[0] = ed[0] + 1
		dd, err := p.DecryptData(ed, k)
		suite.ErrorIs(err, ErrFailedToDecrypt)
		suite.Nil(dd)
	})
}

func (suite *aesGCMSuite) TestEncryptMasterKey() {
	suite.Run("successful encryption", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
	})

	suite.Run("wrong key length", func() {
		p := NewAESGCM()
		mk := []byte("test key")
		mp := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(emk)
	})

	suite.Run("wrong salt length", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := []byte("test master password")
		s := []byte("test salt")
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.ErrorIs(err, ErrInvalidSaltLength)
		suite.Nil(emk)
	})

	suite.Run("nil salt", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := []byte("test master password")
		emk, err := p.EncryptMasterKey(mk, mp, nil)
		suite.ErrorIs(err, ErrInvalidSaltLength)
		suite.Nil(emk)
	})

	suite.Run("nil key", func() {
		p := NewAESGCM()
		mp := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(nil, mp, s)
		suite.ErrorIs(err, ErrInvalidKeyLength)
		suite.Nil(emk)
	})

	suite.Run("nil master password", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, nil, s)
		suite.ErrorIs(err, ErrNoMasterPass)
		suite.Nil(emk)
	})

	suite.Run("empty master password", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := make([]byte, 0)
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.ErrorIs(err, ErrNoMasterPass)
		suite.Nil(emk)
	})
}

func (suite *aesGCMSuite) TestDecryptMasterKey() {
	suite.Run("successful decryption", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		dmk, err := p.DecryptMasterKey(emk, mp, s)
		suite.NoError(err)
		suite.Equal(mk, dmk)
	})

	suite.Run("wrong salt length", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp := []byte("test master password")
		s1, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp, s1)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		s2 := []byte("test salt")
		dmk, err := p.DecryptMasterKey(emk, mp, s2)
		suite.ErrorIs(err, ErrInvalidSaltLength)
		suite.Nil(dmk)
	})

	suite.Run("nil salt", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		s, err := p.NewSalt()
		suite.NoError(err)
		mp := []byte("test master password")
		emk, err := p.EncryptMasterKey(mk, mp, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		dmk, err := p.DecryptMasterKey(emk, mp, nil)
		suite.ErrorIs(err, ErrInvalidSaltLength)
		suite.Nil(dmk)
	})

	suite.Run("wrong master password", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp1 := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp1, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		mp2 := []byte("wrong master password")
		dmk, err := p.DecryptMasterKey(emk, mp2, s)
		suite.ErrorIs(err, ErrFailedToDecrypt)
		suite.Nil(dmk)
	})

	suite.Run("nil master password", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp1 := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp1, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		dmk, err := p.DecryptMasterKey(emk, nil, s)
		suite.ErrorIs(err, ErrNoMasterPass)
		suite.Nil(dmk)
	})

	suite.Run("empty master password", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp1 := []byte("test master password")
		s, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp1, s)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		mp2 := make([]byte, 0)
		dmk, err := p.DecryptMasterKey(emk, mp2, s)
		suite.ErrorIs(err, ErrNoMasterPass)
		suite.Nil(dmk)
	})

	suite.Run("wrong salt", func() {
		p := NewAESGCM()
		mk, err := p.NewKey()
		suite.NoError(err)
		mp1 := []byte("test master password")
		s1, err := p.NewSalt()
		suite.NoError(err)
		emk, err := p.EncryptMasterKey(mk, mp1, s1)
		suite.NoError(err)
		suite.NotEqual(mk, emk)
		s2, err := p.NewSalt()
		suite.NoError(err)
		dmk, err := p.DecryptMasterKey(emk, mp1, s2)
		suite.ErrorIs(err, ErrFailedToDecrypt)
		suite.Nil(dmk)
	})

}
