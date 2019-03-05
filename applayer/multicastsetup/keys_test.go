package multicastsetup

import (
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	mcAddr := lorawan.DevAddr{1, 2, 3, 4}
	mcKey := lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appKey := lorawan.AES128Key{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	genAppKey := lorawan.AES128Key{3, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	mcRootKey := lorawan.AES128Key{4, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	t.Run("GetMcRootKeyForGenAppKey", func(t *testing.T) {
		assert := require.New(t)
		key, err := GetMcRootKeyForGenAppKey(genAppKey)
		assert.NoError(err)
		assert.Equal(lorawan.AES128Key{0x55, 0x34, 0x4e, 0x82, 0x57, 0xe, 0xae, 0xc8, 0xbf, 0x3, 0xb9, 0x99, 0x62, 0xd1, 0xf4, 0x45}, key)
	})

	t.Run("GetMcRootKeyForAppKey", func(t *testing.T) {
		assert := require.New(t)
		key, err := GetMcRootKeyForAppKey(appKey)
		assert.NoError(err)
		assert.Equal(lorawan.AES128Key{0x26, 0x4f, 0xd8, 0x59, 0x58, 0x3f, 0xcc, 0x67, 0x2, 0x41, 0xac, 0x7, 0x1c, 0xc9, 0xf5, 0xbb}, key)
	})

	t.Run("GetMcKEKey", func(t *testing.T) {
		assert := require.New(t)
		key, err := GetMcKEKey(mcRootKey)
		assert.NoError(err)
		assert.Equal(lorawan.AES128Key{0x90, 0x83, 0xbe, 0xbf, 0x70, 0x42, 0x57, 0x88, 0x31, 0x60, 0xdb, 0xfc, 0xde, 0x33, 0xad, 0x71}, key)
	})

	t.Run("GetMcAppSKey", func(t *testing.T) {
		assert := require.New(t)
		key, err := GetMcAppSKey(mcKey, mcAddr)
		assert.NoError(err)
		assert.Equal(lorawan.AES128Key{0x95, 0xcb, 0x45, 0x18, 0xee, 0x37, 0x56, 0x6, 0x73, 0x5b, 0xba, 0xcb, 0xdc, 0xe8, 0x37, 0xfa}, key)
	})

	t.Run("GetMcNetSKey", func(t *testing.T) {
		assert := require.New(t)
		key, err := GetMcNetSKey(mcKey, mcAddr)
		assert.NoError(err)
		assert.Equal(lorawan.AES128Key{0xc3, 0xf6, 0xb3, 0x88, 0xba, 0xd6, 0xc0, 0x0, 0xb2, 0x32, 0x91, 0xad, 0x52, 0xc1, 0x1c, 0x7b}, key)
	})
}
