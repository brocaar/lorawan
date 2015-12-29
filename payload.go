package lorawan

import (
	"encoding"
	"errors"
)

// Payload is the interface that every payload needs to implement.
type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// DataPayload represents a slice of bytes.
type DataPayload struct {
	Bytes []byte
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
	AppEUI   [8]byte
	DevEUI   [8]byte
	DevNonce [2]byte
}

// MarshalBinary marshals the object in binary form.
func (p JoinRequestPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 18)
	b = append(b, p.AppEUI[:]...)
	b = append(b, p.DevEUI[:]...)
	b = append(b, p.DevNonce[:]...)
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *JoinRequestPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 18 {
		return errors.New("lorawan: 18 bytes of data are expected")
	}
	copy(p.AppEUI[:], data[0:8])
	copy(p.DevEUI[:], data[8:16])
	copy(p.DevNonce[:], data[16:18])
	return nil
}

// JoinAcceptPayload represents the join-accept message payload.
// todo: implement CFlist
type JoinAcceptPayload struct {
	AppNonce   [3]byte
	NetID      [3]byte
	DevAddr    DevAddr
	DLSettings DLsettings
	RXDelay    uint8
}

// MarshalBinary marshals the object in binary form.
func (p JoinAcceptPayload) MarshalBinary() ([]byte, error) {
	var b []byte
	var err error
	out := make([]byte, 0, 12)

	out = append(out, p.AppNonce[:]...)
	out = append(out, p.NetID[:]...)

	b, err = p.DevAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	out = append(out, b...)

	b, err = p.DLSettings.MarshalBinary()
	if err != nil {
		return nil, err
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

	copy(p.AppNonce[:], data[0:3])
	copy(p.NetID[:], data[3:6])

	if err := p.DevAddr.UnmarshalBinary(data[6:10]); err != nil {
		return err
	}
	if err := p.DLSettings.UnmarshalBinary(data[10:11]); err != nil {
		return err
	}
	p.RXDelay = uint8(data[11])
	return nil
}
