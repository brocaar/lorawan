package lorawan

import (
	"encoding"
	"encoding/binary"
	"errors"
)

// Payload is the interface that every payload needs to implement.
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Clone() Payload
}

// DataPayload represents a slice of bytes.
type DataPayload struct {
	Bytes []byte
}

// Clone returns a copy of the payload.
func (p DataPayload) Clone() Payload {
	return &p
}

// MarshalBinary marshals the object in binary form.
func (p DataPayload) MarshalBinary() ([]byte, error) {
	return p.Bytes, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *DataPayload) UnmarshalBinary(data []byte) error {
	p.Bytes = make([]byte, len(data))
	copy(p.Bytes, data)
	return nil
}

// JoinRequestPayload represents the join-request message payload.
type JoinRequestPayload struct {
	AppEUI   uint64
	DevEUI   uint64
	DevNonce uint16
}

// Clone returns a copy of the payload.
func (p JoinRequestPayload) Clone() Payload {
	return &p
}

// MarshalBinary marshals the object in binary form.
func (p JoinRequestPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 18)
	binary.LittleEndian.PutUint64(b[0:8], p.AppEUI)
	binary.LittleEndian.PutUint64(b[8:16], p.DevEUI)
	binary.LittleEndian.PutUint16(b[16:18], p.DevNonce)
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *JoinRequestPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 18 {
		return errors.New("lorawan: 18 bytes of data are expected")
	}
	p.AppEUI = binary.LittleEndian.Uint64(data[0:8])
	p.DevEUI = binary.LittleEndian.Uint64(data[8:16])
	p.DevNonce = binary.LittleEndian.Uint16(data[16:18])
	return nil
}
