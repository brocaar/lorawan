package lorawan

import "errors"

// MType represents the message type.
type mType byte

// Major defines the major version of data message.
type major byte

// Supported message types (MType)
const (
	JoinRequest         mType = 0
	JoinAccept          mType = (1 << 5)
	UnconfirmedDataUp   mType = (1 << 6)
	UnconfirmedDataDown mType = (1 << 6) ^ (1 << 5)
	ConfirmedDataUp     mType = (1 << 7)
	ConfirmedDataDown   mType = (1 << 7) ^ (1 << 5)
	Proprietary         mType = (1 << 7) ^ (1 << 6) ^ (1 << 5)
)

// Supported major versions
const (
	LoRaWANR1 major = 0
)

// MHDR represents the MAC header.
type MHDR struct {
	MType mType
	Major major
}

// MarshalBinary marshals the object in binary form.
func (h MHDR) MarshalBinary() ([]byte, error) {
	return []byte{byte(h.Major) ^ byte(h.MType)}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (h *MHDR) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	h.Major = major(data[0] & ((1 << 1) ^ (1 << 0)))
	h.MType = mType(data[0] & ((1 << 7) ^ (1 << 6) ^ (1 << 5)))
	return nil
}
