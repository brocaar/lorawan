//go:generate stringer -type=CID

// Package firmwaremanagement implements the Firmware Management Protocol v1.0.0 over LoRaWAN.
package firmwaremanagement

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// CID defines the command identifier.
type CID byte

// DefaultFPort defines the default fPort value for Firmware Management.
const DefaultFPort uint8 = 203

// Available command identifiers.
const (
	PackageVersionReq     CID = 0x00
	PackageVersionAns     CID = 0x00
	DevVersionReq         CID = 0x01
	DevVersionAns         CID = 0x01
	DevRebootTimeReq      CID = 0x02
	DevRebootTimeAns      CID = 0x02
	DevRebootCountdownReq CID = 0x03
	DevRebootCountdownAns CID = 0x03
	DevUpgradeImageReq    CID = 0x04
	DevUpgradeImageAns    CID = 0x04
	DevDeleteImageReq     CID = 0x05
	DevDeleteImageAns     CID = 0x05
)

// Errors
var (
	ErrNoPayloadForCID = errors.New("lorawan/applayer/firmwaremanagement: no payload for given CID")
)

// map[uplink]...
var commandPayloadRegistry = map[bool]map[CID]func() CommandPayload{
	true: map[CID]func() CommandPayload{
		PackageVersionAns:     func() CommandPayload { return &PackageVersionAnsPayload{} },
		DevVersionAns:         func() CommandPayload { return &DevVersionAnsPayload{} },
		DevRebootTimeAns:      func() CommandPayload { return &DevRebootTimeAnsPayload{} },
		DevRebootCountdownAns: func() CommandPayload { return &DevRebootCountdownAnsPayload{} },
		DevUpgradeImageAns:    func() CommandPayload { return &DevUpgradeImageAnsPayload{} },
		DevDeleteImageAns:     func() CommandPayload { return &DevDeleteImageAnsPayload{} },
	},
	false: map[CID]func() CommandPayload{
		DevVersionReq:         func() CommandPayload { return &DevVersionReqPayload{} },
		DevRebootTimeReq:      func() CommandPayload { return &DevRebootTimeReqPayload{} },
		DevRebootCountdownReq: func() CommandPayload { return &DevRebootCountdownReqPayload{} },
		DevUpgradeImageReq:    func() CommandPayload { return &DevUpgradeImageReqPayload{} },
		DevDeleteImageReq:     func() CommandPayload { return &DevDeleteImageReqPayload{} },
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
		return errors.New("lorawan/applayer/firmwaremanagement: at least 1 byte is expected")
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

// Size returns the payload size in number of bytes.
func (p PackageVersionAnsPayload) Size() int {
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
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.PackageIdentifier = data[0]
	p.PackageVersion = data[1]
	return nil
}

// DevVersionReqPayload implements the DevVersionReq payload.
type DevVersionReqPayload struct{}

// Size returns the payload size in number of bytes.
func (p DevVersionReqPayload) Size() int {
	return 0
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevVersionReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevVersionReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}
	return nil
}

// DevVersionAnsPayload implements the DevVersionAns payload.
type DevVersionAnsPayload struct {
	FWversion uint32
	HWversion uint32
}

// Size returns the payload size in number of bytes.
func (p DevVersionAnsPayload) Size() int {
	return 8
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevVersionAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	binary.LittleEndian.PutUint32(b[0:4], uint32(p.FWversion))
	binary.LittleEndian.PutUint32(b[4:8], uint32(p.HWversion))

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevVersionAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.FWversion = binary.LittleEndian.Uint32(data[0:4])
	p.HWversion = binary.LittleEndian.Uint32(data[4:8])

	return nil
}

// DevRebootTimeReqPayload implements the DevRebootTimeReq payload.
type DevRebootTimeReqPayload struct {
	RebootTime uint32
}

// Size returns the payload size in number of bytes.
func (p DevRebootTimeReqPayload) Size() int {
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevRebootTimeReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	binary.LittleEndian.PutUint32(b[0:4], uint32(p.RebootTime))
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevRebootTimeReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.RebootTime = binary.LittleEndian.Uint32(data[0:4])
	return nil
}

// DevRebootTimeAnsPayload implements the DevRebootTimeAns payload.
type DevRebootTimeAnsPayload struct {
	RebootTime uint32
}

// Size returns the payload size in number of bytes.
func (p DevRebootTimeAnsPayload) Size() int {
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevRebootTimeAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	binary.LittleEndian.PutUint32(b[0:4], uint32(p.RebootTime))
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevRebootTimeAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.RebootTime = binary.LittleEndian.Uint32(data[0:4])

	return nil
}

// DevRebootCountdownReqPayload implements the DevRebootCountdownReq payload.
type DevRebootCountdownReqPayload struct {
	Countdown uint32
}

// Size returns the payload size in number of bytes.
func (p DevRebootCountdownReqPayload) Size() int {
	return 3
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevRebootCountdownReqPayload) MarshalBinary() ([]byte, error) {
	countdownB := make([]byte, 4)
	binary.LittleEndian.PutUint32(countdownB, p.Countdown)
	b := countdownB[:3]

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevRebootCountdownReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	countdownB := make([]byte, 4)
	copy(countdownB, data[0:3])
	p.Countdown = binary.LittleEndian.Uint32(countdownB)

	return nil
}

// DevRebootCountdownAnsPayload implements the DevRebootCountdownAns payload.
type DevRebootCountdownAnsPayload struct {
	Countdown uint32
}

// Size returns the payload size in number of bytes.
func (p DevRebootCountdownAnsPayload) Size() int {
	return 3
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevRebootCountdownAnsPayload) MarshalBinary() ([]byte, error) {
	countdownB := make([]byte, 4)
	binary.LittleEndian.PutUint32(countdownB, p.Countdown)
	b := countdownB[:3]

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevRebootCountdownAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	countdownB := make([]byte, 4)
	copy(countdownB, data[0:3])
	p.Countdown = binary.LittleEndian.Uint32(countdownB)

	return nil
}

// DevUpgradeImageReqPayload implements the DevUpgradeImageReq payload.
type DevUpgradeImageReqPayload struct{}

// Size returns the payload size in number of bytes.
func (p DevUpgradeImageReqPayload) Size() int {
	return 0
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevUpgradeImageReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevUpgradeImageReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}
	return nil
}

// DevUpgradeImageAnsPayload implements the DevUpgradeImageAns payload.
type DevUpgradeImageAnsPayload struct {
	Status              DevUpgradeImageAnsPayloadStatus
	nextFirmwareVersion *uint32
}

// DevUpgradeImageAnsPayloadStatus implements the DevUpgradeImageAnsPayload payload Status field.
type DevUpgradeImageAnsPayloadStatus struct {
	UpImageStatus UpImageStatus
}

type UpImageStatus uint8

const (
	NoFirmwarePresent                 UpImageStatus = 0
	FirmwareCorruptOrInvalidSignature UpImageStatus = 1
	FirmwareIncorrectHardware         UpImageStatus = 2
	FirmwareValid                     UpImageStatus = 3
)

// Size returns the payload size in number of bytes.
func (p DevUpgradeImageAnsPayload) Size() int {
	if p.Status.IsFirmwareImageValid() {
		return 5
	}
	return 1
}

// Is Firmware image valid
func (p DevUpgradeImageAnsPayloadStatus) IsFirmwareImageValid() bool {
	return p.UpImageStatus == FirmwareValid
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevUpgradeImageAnsPayload) MarshalBinary() ([]byte, error) {
	if !p.Status.IsFirmwareImageValid() && p.nextFirmwareVersion != nil {
		return nil, errors.New("lorawan/applayer/firmwaremanagement: nextFirmwareVersion must be nil when UpImageStatus != 3 due no valid firmware present")
	}

	b := make([]byte, p.Size())
	b[0] = uint8(p.Status.UpImageStatus) & 0x3

	if p.Status.IsFirmwareImageValid() {
		binary.LittleEndian.PutUint32(b[1:5], *p.nextFirmwareVersion)
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevUpgradeImageAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 1 {
		return errors.New("lorawan/applayer/firmwaremanagement: at least 1 byte is expected")
	}

	p.Status.UpImageStatus = UpImageStatus(data[0] & 0x3)

	if p.Status.IsFirmwareImageValid() {
		if len(data) < p.Size() {
			return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
		}
		nextFirmareVersion := binary.LittleEndian.Uint32(data[1:5])
		p.nextFirmwareVersion = &nextFirmareVersion
	}

	return nil
}

// DevDeleteImageReqPayload implements the DevDeleteImageReq payload.
type DevDeleteImageReqPayload struct {
	FirmwareToDeleteVersion uint32
}

// Size returns the payload size in number of bytes.
func (p DevDeleteImageReqPayload) Size() int {
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevDeleteImageReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	binary.LittleEndian.PutUint32(b[0:4], p.FirmwareToDeleteVersion)
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevDeleteImageReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.FirmwareToDeleteVersion = binary.LittleEndian.Uint32(data[0:4])

	return nil
}

// DevDeleteImageAnsPayload implements the DevDeleteImageAns payload.
type DevDeleteImageAnsPayload struct {
	Status DevDeleteImageAnsStatus
}

// DevDeleteImageAnsPayloadStatus implements the DevDeleteImageAns payload Status field.
type DevDeleteImageAnsStatus struct {
	ErrorInvalidVersion uint8
	ErrorNoValidImage   uint8
}

// Size returns the payload size in number of bytes.
func (p DevDeleteImageAnsPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p DevDeleteImageAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.Status.ErrorNoValidImage & 0x1
	b[0] = b[0] | (p.Status.ErrorNoValidImage&0x1)<<1

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DevDeleteImageAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/firmwaremanagement: %d bytes are expected", p.Size())
	}

	p.Status.ErrorNoValidImage = data[0] & 0x1
	p.Status.ErrorInvalidVersion = (data[0] >> 1) & 0x1

	return nil
}
