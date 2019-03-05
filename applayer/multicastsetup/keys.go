package multicastsetup

import (
	"crypto/aes"
	"fmt"

	"github.com/brocaar/lorawan"
)

// GetMcRootKeyForGenAppKey returns the McRootKey given a GenAppKey.
// Note: The GenAppKey is only used for LoRaWAN 1.0.x devices.
func GetMcRootKeyForGenAppKey(genAppKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	return getKey(genAppKey, [16]byte{})
}

// GetMcRootKeyForAppKey returns the McRootKey given an AppKey.
// Note: The AppKey is only used for LoRaWAN 1.1.x devices.
func GetMcRootKeyForAppKey(appKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	return getKey(appKey, [16]byte{0x20})
}

// GetMcKEKey returns the McKEKey given the McRootKey.
func GetMcKEKey(mcRootKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	return getKey(mcRootKey, [16]byte{})
}

// GetMcAppSKey returns the McAppSKey given the McKey and McAddr.
func GetMcAppSKey(mcKey lorawan.AES128Key, mcAddr lorawan.DevAddr) (lorawan.AES128Key, error) {
	b := [16]byte{0x01}

	mcAddrB, err := mcAddr.MarshalBinary()
	if err != nil {
		return lorawan.AES128Key{}, err
	}
	copy(b[1:5], mcAddrB)

	return getKey(mcKey, b)
}

// GetMcNetSKey returns the McNetSKey given the McKey and McAddr.
func GetMcNetSKey(mcKey lorawan.AES128Key, mcAddr lorawan.DevAddr) (lorawan.AES128Key, error) {
	b := [16]byte{0x02}

	mcAddrB, err := mcAddr.MarshalBinary()
	if err != nil {
		return lorawan.AES128Key{}, err
	}
	copy(b[1:5], mcAddrB)

	return getKey(mcKey, b)
}

func getKey(key lorawan.AES128Key, b [16]byte) (lorawan.AES128Key, error) {
	var out lorawan.AES128Key

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return out, err
	}
	if block.BlockSize() != len(b) {
		return out, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}

	block.Encrypt(out[:], b[:])
	return out, nil
}
