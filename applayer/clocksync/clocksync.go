// Package clocksync implements the Application Layer Clock Synchronization v1.0.0 over LoRaWAN.
package clocksync

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// CID defines the command identifier.
type CID byte

// DefaultFPort defines the default fPort value for Clock Synnchronization.
const DefaultFPort uint8 = 202

// Available command identifiers.
const (
	PackageVersionReq           CID = 0x00
	PackageVersionAns           CID = 0x00
	AppTimeReq                  CID = 0x01
	AppTimeAns                  CID = 0x01
	DeviceAppTimePeriodicityReq CID = 0x02
	DeviceAppTimePeriodicityAns CID = 0x02
	ForceDeviceResyncReq        CID = 0x03
)

type commandPayloadInfo struct {
	size    int // the number of payload bytes
	payload func() CommandPayload
}

// map[uplink]...
var commandPayloadRegistry = map[bool]map[CID]commandPayloadInfo{
	true: map[CID]commandPayloadInfo{
		PackageVersionAns:           {2, func() CommandPayload { return &PackageVersionAnsPayload{} }},
		AppTimeReq:                  {5, func() CommandPayload { return &AppTimeReqPayload{} }},
		DeviceAppTimePeriodicityAns: {5, func() CommandPayload { return &DeviceAppTimePeriodicityAnsPayload{} }},
	},
	false: map[CID]commandPayloadInfo{
		AppTimeAns:                  {5, func() CommandPayload { return &AppTimeAnsPayload{} }},
		DeviceAppTimePeriodicityReq: {1, func() CommandPayload { return &DeviceAppTimePeriodicityReqPayload{} }},
		ForceDeviceResyncReq:        {1, func() CommandPayload { return &ForceDeviceResyncReqPayload{} }},
	},
}

// GetCommandPayloadAndSize returns a new CommandPayload and its size.
func GetCommandPayloadAndSize(uplink bool, c CID) (CommandPayload, int, error) {
	v, ok := commandPayloadRegistry[uplink][c]
	if !ok {
		return nil, 0, fmt.Errorf("lorawan/applayer/clocksync: payload unknown for uplink: %v and CID=%v", uplink, c)
	}

	return v.payload(), v.size, nil
}

// CommandPayload defines the interface that a command payload must implement.
type CommandPayload interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

// Command defines the Command structure.
type Command struct {
	CID     CID
	Payload CommandPayload
}

// MarshalBinary encodes the command to a slice of bytes.
func (c Command) MarshalBinary() ([]byte, error) {
	b := []byte{byte(c.CID)}

	if c.Payload != nil {
		p, err := c.Payload.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b = append(b, p...)
	}

	return b, nil
}

// UnmarshalBinary decodes a slice of bytes into a command.
func (c *Command) UnmarshalBinary(uplink bool, data []byte) error {
	if len(data) == 0 {
		return errors.New("lorawan/applayer/clocksync: at least 1 byte is expected")
	}

	c.CID = CID(data[0])

	if len(data) > 1 {
		p, _, err := GetCommandPayloadAndSize(uplink, c.CID)
		if err != nil {
			return err
		}
		c.Payload = p
		if err := c.Payload.UnmarshalBinary(data[1:]); err != nil {
			return err
		}
	}

	return nil
}

// PackageVersionAnsPayload implements the PackageVersionAns payload.
type PackageVersionAnsPayload struct {
	PackageIdentifier uint8
	PackageVersion    uint8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p PackageVersionAnsPayload) MarshalBinary() ([]byte, error) {
	return []byte{
		p.PackageIdentifier,
		p.PackageVersion,
	}, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *PackageVersionAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("lorawan/applayer/clocksync: exactly 2 bytes are expected")
	}

	p.PackageIdentifier = data[0]
	p.PackageVersion = data[1]
	return nil
}

// AppTimeReqPayload implements the AppTimeReq payload.
type AppTimeReqPayload struct {
	DeviceTime uint32
	Param      AppTimeReqPayloadParam
}

// AppTimeReqParam implements the AppTimeReq Param field.
type AppTimeReqPayloadParam struct {
	AnsRequired bool
	TokenReq    uint8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p AppTimeReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)

	binary.LittleEndian.PutUint32(b[0:4], p.DeviceTime)
	b[4] = p.Param.TokenReq & 0x0f // only the first 4 bytes
	if p.Param.AnsRequired {
		b[4] |= 1 << 4
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *AppTimeReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("lorawan/applayer/clocksync: exactly 5 bytes are expected")
	}

	p.DeviceTime = binary.LittleEndian.Uint32(data[0:4])
	p.Param.TokenReq = uint8(data[4] & 0x0f)
	if data[4]&(1<<4) != 0 {
		p.Param.AnsRequired = true
	}

	return nil
}

// AppTimeAnsPayload implements the AppTimeAns payload.
type AppTimeAnsPayload struct {
	TimeCorrection int32
	Param          AppTimeAnsPayloadParam
}

// AppTimeAnsPayloadParam implements the AppTimeAns payload Param field.
type AppTimeAnsPayloadParam struct {
	TokenAns uint8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p AppTimeAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)

	binary.LittleEndian.PutUint32(b[0:4], uint32(p.TimeCorrection))
	b[4] = p.Param.TokenAns & 0x0f // only the first 4 bytes

	return b, nil
}

// UnmarshalBinary decoces the payload from a slice of bytes.
func (p *AppTimeAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("lorawan/applayer/clocksync: exactly 5 bytes are expected")
	}

	p.TimeCorrection = int32(binary.LittleEndian.Uint32(data[0:4]))
	p.Param.TokenAns = uint8(data[4] & 0x0f)

	return nil
}

// DeviceAppTimePeriodicityReqPayload implements the DeviceAppTimePeriodicityReq payload.
type DeviceAppTimePeriodicityReqPayload struct {
	Periodicity DeviceAppTimePeriodicityReqPayloadPeriodicity
}

// DeviceAppTimePeriodicityReqPayloadPeriodicity implements the DeviceAppTimePeriodicityReq payload Periodicity field.
type DeviceAppTimePeriodicityReqPayloadPeriodicity struct {
	Period uint8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DeviceAppTimePeriodicityReqPayload) MarshalBinary() ([]byte, error) {
	out := make([]byte, 1)
	out[0] = p.Periodicity.Period & 0x0f // only the first 4 bytes

	return out, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DeviceAppTimePeriodicityReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan/applayer/clocksync: exactly 1 byte is expected")
	}

	p.Periodicity.Period = data[0] & 0x0f

	return nil
}

// DeviceAppTimePeriodicityAnsPayload implements the DeviceAppTimePeriodicityAns payload.
type DeviceAppTimePeriodicityAnsPayload struct {
	Status DeviceAppTimePeriodicityAnsPayloadStatus
	Time   uint32
}

// DeviceAppTimePeriodicityAnsStatus implements the DeviceAppTimePeriodicityAns status field.
type DeviceAppTimePeriodicityAnsPayloadStatus struct {
	NotSupported bool
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DeviceAppTimePeriodicityAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)
	if p.Status.NotSupported {
		b[0] = 1
	}
	binary.LittleEndian.PutUint32(b[1:5], p.Time)
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DeviceAppTimePeriodicityAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("lorawan/applayer/clocksync: exactly 5 bytes are expected")
	}
	if data[0]&1 != 0 {
		p.Status.NotSupported = true
	}
	p.Time = binary.LittleEndian.Uint32(data[1:5])
	return nil
}

// ForceDeviceResyncReqPayload implements the ForceDeviceResyncReq payload.
type ForceDeviceResyncReqPayload struct {
	ForceConf ForceDeviceResyncReqPayloadForceConf
}

// ForceDeviceResyncReqPayloadForceConf implements the ForceDeviceResyncReq payload ForceConf field.
type ForceDeviceResyncReqPayloadForceConf struct {
	NbTransmissions uint8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p ForceDeviceResyncReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = p.ForceConf.NbTransmissions & 0x17 // first 3 bits
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *ForceDeviceResyncReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan/applayer/clocksync: exactly 1 byte is expected")
	}
	p.ForceConf.NbTransmissions = data[0] & 0x17
	return nil
}
