package lorawan

import "encoding"

// Payload is the interface that every payload needs to implement.
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	Clone() Payload
}

// DataPayload represents
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
