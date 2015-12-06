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

// JoinAcceptPayload represents the join-accept message payload.
// todo: implement CFlist
type JoinAcceptPayload struct {
	AppNonce   uint32 // only a value of up to 24 bits can be used
	NetID      uint32 // only a value of up to 24 bits can be used
	DevAddr    DevAddr
	DLSettings DLsettings
	RXDelay    uint8
}

// Clone returns a copy of the payload.
func (p JoinAcceptPayload) Clone() Payload {
	return &p
}

// MarshalBinary marshals the object in binary form.
func (p JoinAcceptPayload) MarshalBinary() ([]byte, error) {
	if p.AppNonce > 16777216 { // 2^24
		return []byte{}, errors.New("lorawan: max value of AppNonce is 2^24 - 1")
	}
	if p.NetID > 16777216 {
		return []byte{}, errors.New("lorawan: max value of NetID is 2^24 - 1")
	}

	var b []byte
	var err error
	out := make([]byte, 7, 12)
	binary.LittleEndian.PutUint32(out[0:4], p.AppNonce)
	binary.LittleEndian.PutUint32(out[3:7], uint32(p.NetID))
	out = out[0:6] // drop the last empty byte (NetID is only 24 bit, so 8 MSB are 0)

	b, err = p.DevAddr.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	out = append(out, b...)

	b, err = p.DLSettings.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	out = append(out, b...)
	out = append(out, byte(p.RXDelay))

	return out, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *JoinAcceptPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 12 {
		return errors.New("lorawan: 12 bytes of data are expected")
	}
	i := make([]byte, 4) // Uint32 expects 4 bytes of data
	copy(i, data[0:3])
	p.AppNonce = binary.LittleEndian.Uint32(i)
	copy(i, data[3:6])
	p.NetID = binary.LittleEndian.Uint32(i)
	if err := p.DevAddr.UnmarshalBinary(data[6:10]); err != nil {
		return err
	}
	if err := p.DLSettings.UnmarshalBinary(data[10:11]); err != nil {
		return err
	}
	p.RXDelay = uint8(data[11])
	return nil
}
