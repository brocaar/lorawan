//go:generate stringer -type=CID

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

// Errors
var (
	ErrNoPayloadForCID = errors.New("lorawan/applayer/clocksync: no payload for given CID")
)

// map[uplink]...
var commandPayloadRegistry = map[bool]map[CID]func() CommandPayload{
	true: map[CID]func() CommandPayload{
		PackageVersionAns:           func() CommandPayload { return &PackageVersionAnsPayload{} },
		AppTimeReq:                  func() CommandPayload { return &AppTimeReqPayload{} },
		DeviceAppTimePeriodicityAns: func() CommandPayload { return &DeviceAppTimePeriodicityAnsPayload{} },
	},
	false: map[CID]func() CommandPayload{
		AppTimeAns:                  func() CommandPayload { return &AppTimeAnsPayload{} },
		DeviceAppTimePeriodicityReq: func() CommandPayload { return &DeviceAppTimePeriodicityReqPayload{} },
		ForceDeviceResyncReq:        func() CommandPayload { return &ForceDeviceResyncReqPayload{} },
	},
}

// GetCommandPayload returns a new CommandPayload for the given CID.
func GetCommandPayload(uplink bool, c CID) (CommandPayload, error) {
	v, ok := commandPayloadRegistry[uplink][c]
	if !ok {
		return nil, ErrNoPayloadForCID
	}

	return v(), nil
}

// CommandPayload defines the interface that a command payload must implement.
type CommandPayload interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
	Size() int
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

	p, err := GetCommandPayload(uplink, c.CID)
	if err != nil {
		if err == ErrNoPayloadForCID {
			return nil
		}
		return err
	}

	c.Payload = p
	if err := c.Payload.UnmarshalBinary(data[1:]); err != nil {
		return err
	}

	return nil
}

// Size returns the size of the command in bytes.
func (c Command) Size() int {
	if c.Payload != nil {
		return c.Payload.Size() + 1
	}
	return 1
}

// Commands defines a slice of commands.
type Commands []Command

// MarshalBinary encodes the commands to a slice of bytes.
func (c Commands) MarshalBinary() ([]byte, error) {
	var out []byte

	for _, cmd := range c {
		b, err := cmd.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out = append(out, b...)
	}
	return out, nil
}

// UnmarshalBinary decodes a slice of bytes into a slice of commands.
func (c *Commands) UnmarshalBinary(uplink bool, data []byte) error {
	var i int

	for i < len(data) {
		var cmd Command
		if err := cmd.UnmarshalBinary(uplink, data[i:]); err != nil {
			return err
		}
		i += cmd.Size()
		*c = append(*c, cmd)
	}

	return nil
}

// PackageVersionAnsPayload implements the PackageVersionAns payload.
type PackageVersionAnsPayload struct {
	PackageIdentifier uint8
	PackageVersion    uint8
}

// Size returns the payload size in bytes.
func (p *PackageVersionAnsPayload) Size() int {
	return 2
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
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
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

// AppTimeReqPayloadParam implements the AppTimeReq Param field.
type AppTimeReqPayloadParam struct {
	AnsRequired bool
	TokenReq    uint8
}

// Size returns the payload size in bytes.
func (p AppTimeReqPayload) Size() int {
	return 5
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p AppTimeReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	binary.LittleEndian.PutUint32(b[0:4], p.DeviceTime)
	b[4] = p.Param.TokenReq & 0x0f // only the first 4 bytes
	if p.Param.AnsRequired {
		b[4] |= 1 << 4
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *AppTimeReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
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

// Size returns the payload size in bytes.
func (p *AppTimeAnsPayload) Size() int {
	return 5
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p AppTimeAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	binary.LittleEndian.PutUint32(b[0:4], uint32(p.TimeCorrection))
	b[4] = p.Param.TokenAns & 0x0f // only the first 4 bytes

	return b, nil
}

// UnmarshalBinary decoces the payload from a slice of bytes.
func (p *AppTimeAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
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

// Size returns the payload size in bytes.
func (p DeviceAppTimePeriodicityReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DeviceAppTimePeriodicityReqPayload) MarshalBinary() ([]byte, error) {
	out := make([]byte, p.Size())
	out[0] = p.Periodicity.Period & 0x0f // only the first 4 bytes

	return out, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DeviceAppTimePeriodicityReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
	}

	p.Periodicity.Period = data[0] & 0x0f

	return nil
}

// DeviceAppTimePeriodicityAnsPayload implements the DeviceAppTimePeriodicityAns payload.
type DeviceAppTimePeriodicityAnsPayload struct {
	Status DeviceAppTimePeriodicityAnsPayloadStatus
	Time   uint32
}

// DeviceAppTimePeriodicityAnsPayloadStatus implements the DeviceAppTimePeriodicityAns status field.
type DeviceAppTimePeriodicityAnsPayloadStatus struct {
	NotSupported bool
}

// Size returns the payload size in bytes.
func (p DeviceAppTimePeriodicityAnsPayload) Size() int {
	return 5
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DeviceAppTimePeriodicityAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	if p.Status.NotSupported {
		b[0] = 1
	}
	binary.LittleEndian.PutUint32(b[1:5], p.Time)
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DeviceAppTimePeriodicityAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
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

// Size returns the payload size in bytes.
func (p ForceDeviceResyncReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p ForceDeviceResyncReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.ForceConf.NbTransmissions & 0x17 // first 3 bits
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *ForceDeviceResyncReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/clocksync: %d bytes are expected", p.Size())
	}
	p.ForceConf.NbTransmissions = data[0] & 0x17
	return nil
}
