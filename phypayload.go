package lorawan

import (
	"encoding/binary"
	"errors"

	"github.com/jacobsa/crypto/cmac"
)

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

// PHYPayload represents the physical payload.
type PHYPayload struct {
	MHDR       MHDR
	MACPayload MACPayload
	MIC        [4]byte
	uplink     bool
}

// New returns a new PHYPayload instance set to either uplink or downlink.
func New(uplink bool) PHYPayload {
	return PHYPayload{uplink: uplink}
}

// calculateMIC calculates and returns the MIC.
func (p PHYPayload) calculateMIC(nwkSKey []byte) ([]byte, error) {
	var b []byte
	var err error
	micBytes := make([]byte, 0)

	b, err = p.MHDR.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	b, err = p.MACPayload.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	b0 := make([]byte, 16)
	b0[0] = 0x49
	if p.uplink {
		b0[5] = 1
	}
	binary.LittleEndian.PutUint32(b0[6:10], uint32(p.MACPayload.FHDR.DevAddr))
	binary.LittleEndian.PutUint32(b0[10:14], uint32(p.MACPayload.FHDR.Fcnt))
	b0[15] = byte(len(micBytes))

	hash, err := cmac.New(nwkSKey)
	if err != nil {
		return []byte{}, err
	}

	hash.Write(b0)
	hash.Write(micBytes)

	hb := hash.Sum([]byte{})
	if len(hb) < 4 {
		return []byte{}, errors.New("lorawan: the hash returned less than 4 bytes")
	}
	return hb[0:4], nil
}

// SetMIC calculates and sets the MIC field.
func (p *PHYPayload) SetMIC(nwkSKey []byte) error {
	mic, err := p.calculateMIC(nwkSKey)
	if err != nil {
		return err
	}
	if len(mic) != 4 {
		return errors.New("lorawan: a MIC of 4 bytes is expected")
	}
	for i, v := range mic {
		p.MIC[i] = v
	}
	return nil
}

// ValidateMIC returns if the MIC is valid.
func (p PHYPayload) ValidateMIC(nwkSKey []byte) (bool, error) {
	mic, err := p.calculateMIC(nwkSKey)
	if err != nil {
		return false, err
	}
	if len(mic) != 4 {
		return false, errors.New("lorawan: a MIC of 4 bytes is expected")
	}
	for i, v := range mic {
		if p.MIC[i] != v {
			return false, nil
		}
	}
	return true, nil
}

// MarshalBinary marshals the object in binary form.
func (p PHYPayload) MarshalBinary() ([]byte, error) {
	var out []byte
	var b []byte
	var err error

	if b, err = p.MHDR.MarshalBinary(); err != nil {
		return []byte{}, err
	}
	out = append(out, b...)

	p.MACPayload.uplink = p.uplink
	if b, err = p.MACPayload.MarshalBinary(); err != nil {
		return []byte{}, err
	}
	out = append(out, b...)
	out = append(out, p.MIC[0:len(p.MIC)]...)
	return out, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *PHYPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 12 {
		return errors.New("lorawan: at least 12 bytes needed to decode PHYPayload")
	}

	if err := p.MHDR.UnmarshalBinary(data[0:1]); err != nil {
		return err
	}
	p.MACPayload.uplink = p.uplink
	if err := p.MACPayload.UnmarshalBinary(data[1 : len(data)-4]); err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		p.MIC[i] = data[len(data)-4+i]
	}
	return nil
}
