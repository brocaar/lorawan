package lorawan

import (
	"encoding/binary"
	"errors"
)

// DevAddr represents the device address.
type DevAddr uint32

// MarshalBinary marshals the object in binary form.
func (a DevAddr) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(a))
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (a *DevAddr) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}
	*a = DevAddr(binary.LittleEndian.Uint32(data))
	return nil
}

// FCtrl represents the FCtrl (frame control) field.
type FCtrl struct {
	ADR       bool
	ADRACKReq bool
	ACK       bool
	FPending  bool // unly used for downlink messages
	FOptsLen  uint8
}

// MarshalBinary marshals the object in binary form.
func (c FCtrl) MarshalBinary() ([]byte, error) {
	if c.FOptsLen > 15 {
		return []byte{}, errors.New("lorawan: max value of FOptsLen is 15")
	}
	b := byte(c.FOptsLen)
	if c.FPending {
		b = b ^ (1 << 4)
	}
	if c.ACK {
		b = b ^ (1 << 5)
	}
	if c.ADRACKReq {
		b = b ^ (1 << 6)
	}
	if c.ADR {
		b = b ^ (1 << 7)
	}
	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (c *FCtrl) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	c.FOptsLen = data[0] & ((1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0))
	c.FPending = data[0]&(1<<4) > 0
	c.ACK = data[0]&(1<<5) > 0
	c.ADRACKReq = data[0]&(1<<6) > 0
	c.ADR = data[0]&(1<<7) > 0
	return nil
}

// FHDR represents the frame header.
type FHDR struct {
	DevAddr DevAddr
	FCtrl   FCtrl
	Fcnt    uint16
	FOpts   []byte // max. number of allowed bytes is 15
}
