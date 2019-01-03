package joinserver

import (
	"crypto/aes"
	"fmt"

	"github.com/pkg/errors"

	"github.com/brocaar/lorawan"
)

// getFNwkSIntKey returns the FNwkSIntKey.
// For LoRaWAN 1.0: SNwkSIntKey = NwkSEncKey = FNwkSIntKey = NwkSKey
func getFNwkSIntKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x01, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getAppSKey returns appSKey.
func getAppSKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x02, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getSNwkSIntKey returns the NwkSIntKey.
func getSNwkSIntKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x03, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getNwkSEncKey returns the NwkSEncKey.
func getNwkSEncKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x04, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

// getJSIntKey returns the JSIntKey.
func getJSIntKey(nwkKey lorawan.AES128Key, devEUI lorawan.EUI64) (lorawan.AES128Key, error) {
	return getJSKey(0x06, devEUI, nwkKey)
}

// getJSEncKey returns the JSEncKey.
func getJSEncKey(nwkKey lorawan.AES128Key, devEUI lorawan.EUI64) (lorawan.AES128Key, error) {
	return getJSKey(0x05, devEUI, nwkKey)
}

func getSKey(optNeg bool, typ byte, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)
	b[0] = typ

	netIDB, err := netID.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinEUIB, err := joinEUI.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinNonceB, err := joinNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	devNonceB, err := devNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	if optNeg {
		copy(b[1:4], joinNonceB)
		copy(b[4:12], joinEUIB)
		copy(b[12:14], devNonceB)
	} else {
		copy(b[1:4], joinNonceB)
		copy(b[4:7], netIDB)
		copy(b[7:9], devNonceB)
	}

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)

	return key, nil
}

func getJSKey(typ byte, devEUI lorawan.EUI64, nwkKey lorawan.AES128Key) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)

	b[0] = typ

	devB, err := devEUI.MarshalBinary()
	if err != nil {
		return key, err
	}
	copy(b[1:9], devB[:])

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)
	return key, nil
}
