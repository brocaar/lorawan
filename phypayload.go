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
	MACPayload Payload
	MIC        [4]byte
	uplink     bool
}

// New returns a new PHYPayload instance set to either uplink or downlink.
func New(uplink bool) PHYPayload {
	return PHYPayload{uplink: uplink}
}

// calculateMIC calculates and returns the MIC.
func (p PHYPayload) calculateMIC(key []byte) ([]byte, error) {
	if p.MACPayload == nil {
		return []byte{}, errors.New("lorawan: MACPayload should not be empty")
	}

	macPayload, ok := p.MACPayload.(*MACPayload)
	if !ok {
		return []byte{}, errors.New("lorawan: MACPayload should be of type *MACPayload")
	}

	var b []byte
	var err error
	var micBytes []byte

	b, err = p.MHDR.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	b, err = macPayload.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	b0 := make([]byte, 16)
	b0[0] = 0x49
	if p.uplink {
		b0[5] = 1
	}
	binary.LittleEndian.PutUint32(b0[6:10], uint32(macPayload.FHDR.DevAddr))
	binary.LittleEndian.PutUint32(b0[10:14], uint32(macPayload.FHDR.Fcnt))
	b0[15] = byte(len(micBytes))

	hash, err := cmac.New(key)
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

// calculateJoinRequestMIC calculates and returns the join-request MIC.
func (p PHYPayload) calculateJoinRequestMIC(key []byte) ([]byte, error) {
	if p.MACPayload == nil {
		return []byte{}, errors.New("lorawan: MACPayload should not be empty")
	}
	jrPayload, ok := p.MACPayload.(*JoinRequestPayload)
	if !ok {
		return []byte{}, errors.New("lorawan: MACPayload should be of type *JoinRequestPayload")
	}

	var b []byte
	var err error
	micBytes := make([]byte, 0, 19)
	iBytes := make([]byte, 8)

	b, err = p.MHDR.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	binary.LittleEndian.PutUint64(iBytes, jrPayload.AppEUI)
	micBytes = append(micBytes, iBytes...)
	binary.LittleEndian.PutUint64(iBytes, jrPayload.DevEUI)
	micBytes = append(micBytes, iBytes...)
	binary.LittleEndian.PutUint16(iBytes[0:2], jrPayload.DevNonce)
	micBytes = append(micBytes, iBytes[0:2]...)

	hash, err := cmac.New(key)
	if err != nil {
		return []byte{}, err
	}
	hash.Write(micBytes)
	hb := hash.Sum([]byte{})
	if len(hb) < 4 {
		return []byte{}, errors.New("lorawan: the hash returned less than 4 bytes")
	}
	return hb[0:4], nil
}

// calculateJoinAcceptMIC calculates and returns the join-accept MIC.
// todo: is RFU an empty byte or just an empty placeholder (no byte(s))?
func (p PHYPayload) calculateJoinAcceptMIC(key []byte) ([]byte, error) {
	if p.MACPayload == nil {
		return []byte{}, errors.New("lorawan: MACPayload should not be empty")
	}
	jaPayload, ok := p.MACPayload.(*JoinAcceptPayload)
	if !ok {
		return []byte{}, errors.New("lorawan: MACPayload should be of type *JoinAcceptPayload")
	}

	var b []byte
	var err error
	micBytes := make([]byte, 0, 13)
	iBytes := make([]byte, 4)

	b, err = p.MHDR.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)

	binary.LittleEndian.PutUint32(iBytes, jaPayload.AppNonce)
	micBytes = append(micBytes, b[0:3]...)
	binary.LittleEndian.PutUint32(iBytes, jaPayload.NetID)
	micBytes = append(micBytes, b[0:3]...)
	binary.LittleEndian.PutUint32(iBytes, uint32(jaPayload.DevAddr))
	micBytes = append(micBytes, b...)

	b, err = jaPayload.DLSettings.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	micBytes = append(micBytes, b...)
	micBytes = append(micBytes, byte(jaPayload.RXDelay))

	hash, err := cmac.New(key)
	if err != nil {
		return []byte{}, err
	}
	hash.Write(micBytes)
	hb := hash.Sum([]byte{})
	if len(hb) < 4 {
		return []byte{}, errors.New("lorawan: the hash returned less than 4 bytes")
	}
	return hb[0:4], nil
}

// SetMIC calculates and sets the MIC field.
func (p *PHYPayload) SetMIC(key []byte) error {
	var mic []byte
	var err error

	switch p.MACPayload.(type) {
	case *JoinRequestPayload:
		mic, err = p.calculateJoinRequestMIC(key)
	case *JoinAcceptPayload:
		mic, err = p.calculateJoinAcceptMIC(key)
	default:
		mic, err = p.calculateMIC(key)
	}

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
func (p PHYPayload) ValidateMIC(key []byte) (bool, error) {
	var mic []byte
	var err error

	switch p.MACPayload.(type) {
	case *JoinRequestPayload:
		mic, err = p.calculateJoinRequestMIC(key)
	case *JoinAcceptPayload:
		mic, err = p.calculateJoinAcceptMIC(key)
	default:
		mic, err = p.calculateMIC(key)
	}

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
	if p.MACPayload == nil {
		return []byte{}, errors.New("lorawan: MACPayload should not be nil")
	}

	if mpl, ok := p.MACPayload.(*MACPayload); ok {
		mpl.uplink = p.uplink
	}

	var out []byte
	var b []byte
	var err error

	if b, err = p.MHDR.MarshalBinary(); err != nil {
		return []byte{}, err
	}
	out = append(out, b...)

	if b, err = p.MACPayload.MarshalBinary(); err != nil {
		return []byte{}, err
	}
	out = append(out, b...)
	out = append(out, p.MIC[0:len(p.MIC)]...)
	return out, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *PHYPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 5 {
		return errors.New("lorawan: at least 5 bytes needed to decode PHYPayload")
	}

	if err := p.MHDR.UnmarshalBinary(data[0:1]); err != nil {
		return err
	}

	switch p.MHDR.MType {
	case JoinRequest:
		p.MACPayload = &JoinRequestPayload{}
	case JoinAccept:
		p.MACPayload = &JoinAcceptPayload{}
	default:
		p.MACPayload = &MACPayload{uplink: p.uplink}
	}

	if err := p.MACPayload.UnmarshalBinary(data[1 : len(data)-4]); err != nil {
		return err
	}
	for i := 0; i < 4; i++ {
		p.MIC[i] = data[len(data)-4+i]
	}
	return nil
}
