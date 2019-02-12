//go:generate stringer -type=CID

// Package fragmentation implements the Fragmented Data Block Transport v1.0.0 over LoRaWAN.
package fragmentation

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// CID defines the command identifier.
type CID byte

// DefaultFPort defines the default fPort value for Fragmented Data Block Transport.
const DefaultFPort uint8 = 201

// Available command identifier.
const (
	PackageVersionReq    CID = 0x00
	PackageVersionAns    CID = 0x00
	FragSessionStatusReq CID = 0x01
	FragSessionStatusAns CID = 0x01
	FragSessionSetupReq  CID = 0x02
	FragSessionSetupAns  CID = 0x02
	FragSessionDeleteReq CID = 0x03
	FragSessionDeleteAns CID = 0x03
	DataFragment         CID = 0x08
)

// Errors
var (
	ErrNoPayloadForCID = errors.New("lorawan/applayer/fragmentation: no payload for given CID")
)

// map[uplink]...
var commandPayloadRegistry = map[bool]map[CID]func() CommandPayload{
	true: map[CID]func() CommandPayload{
		PackageVersionAns:    func() CommandPayload { return &PackageVersionAnsPayload{} },
		FragSessionSetupAns:  func() CommandPayload { return &FragSessionSetupAnsPayload{} },
		FragSessionDeleteAns: func() CommandPayload { return &FragSessionDeleteAnsPayload{} },
		FragSessionStatusAns: func() CommandPayload { return &FragSessionStatusAnsPayload{} },
	},
	false: map[CID]func() CommandPayload{
		FragSessionSetupReq:  func() CommandPayload { return &FragSessionSetupReqPayload{} },
		FragSessionDeleteReq: func() CommandPayload { return &FragSessionDeleteReqPayload{} },
		DataFragment:         func() CommandPayload { return &DataFragmentPayload{} },
		FragSessionStatusReq: func() CommandPayload { return &FragSessionStatusReqPayload{} },
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
		return errors.New("lorawan/applayer/fragmentation: at least 1 byte is expected")
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
		return fmt.Errorf("lorawan/applayer/fragmentation: %d bytes are expected", p.Size())
	}

	p.PackageIdentifier = data[0]
	p.PackageVersion = data[1]
	return nil
}

// FragSessionSetupReqPayload implements the FragSessionSetupReq payload.
type FragSessionSetupReqPayload struct {
	FragSession FragSessionSetupReqPayloadFragSession
	NbFrag      uint16
	FragSize    uint8
	Control     FragSessionSetupReqPayloadControl
	Padding     uint8
	Descriptor  [4]byte
}

// FragSessionSetupReqPayloadFragSession implements the FragSessionSetupReq payload FragSession field.
type FragSessionSetupReqPayloadFragSession struct {
	FragIndex      uint8
	McGroupBitMask [4]bool
}

// FragSessionSetupReqPayloadControl implements the FragSessionSetupReq payload Control field.
type FragSessionSetupReqPayloadControl struct {
	FragmentationMatrix uint8
	BlockAckDelay       uint8
}

// Size returns the payload size in number of bytes.
func (p FragSessionSetupReqPayload) Size() int {
	return 10
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionSetupReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	// FragSession
	for i, mask := range p.FragSession.McGroupBitMask {
		if mask {
			b[0] |= (1 << uint8(i))
		}
	}
	b[0] |= (p.FragSession.FragIndex & 0x03) << 4

	// NbFrag
	binary.LittleEndian.PutUint16(b[1:3], p.NbFrag)

	// FragSize
	b[3] = p.FragSize

	// Control
	b[4] = p.Control.BlockAckDelay & 0x07
	b[4] |= (p.Control.FragmentationMatrix & 0x07) << 3

	// Padding
	b[5] = p.Padding

	// Descriptor
	copy(b[6:10], p.Descriptor[:])

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d bytes are expected", p.Size())
	}

	// Fragmentation
	for i := range p.FragSession.McGroupBitMask {
		p.FragSession.McGroupBitMask[i] = data[0]&(1<<uint8(i)) != 0
	}
	p.FragSession.FragIndex = (data[0] >> 4) & 0x03

	// NbFrag
	p.NbFrag = binary.LittleEndian.Uint16(data[1:3])

	// FragSize
	p.FragSize = data[3]

	// Control
	p.Control.BlockAckDelay = data[4] & 0x07
	p.Control.FragmentationMatrix = (data[4] >> 3) & 0x07

	// Padding
	p.Padding = data[5]

	// Descriptor
	copy(p.Descriptor[:], data[6:10])

	return nil
}

// FragSessionSetupAnsPayload implements the FragSessionSetupAns payload.
type FragSessionSetupAnsPayload struct {
	StatusBitMask FragSessionSetupAnsPayloadStatusBitMask
}

// FragSessionSetupAnsPayloadStatusBitMask implements the FragSessionSetupAns payload StatusBitMask field.
type FragSessionSetupAnsPayloadStatusBitMask struct {
	FragIndex                    uint8
	WrongDescriptor              bool
	FragSessionIndexNotSupported bool
	NotEnoughMemory              bool
	EncodingUnsupported          bool
}

// Size returns the paylaod size in bytes.
func (p FragSessionSetupAnsPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionSetupAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	if p.StatusBitMask.EncodingUnsupported {
		b[0] |= 0x01
	}

	if p.StatusBitMask.NotEnoughMemory {
		b[0] |= 0x02
	}

	if p.StatusBitMask.FragSessionIndexNotSupported {
		b[0] |= 0x04
	}

	if p.StatusBitMask.WrongDescriptor {
		b[0] |= 0x08
	}

	b[0] |= (p.StatusBitMask.FragIndex & 0x03) << 6

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionSetupAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d byte is expected", p.Size())
	}

	p.StatusBitMask.EncodingUnsupported = data[0]&0x01 != 0
	p.StatusBitMask.NotEnoughMemory = data[0]&0x02 != 0
	p.StatusBitMask.FragSessionIndexNotSupported = data[0]&0x04 != 0
	p.StatusBitMask.WrongDescriptor = data[0]&0x08 != 0
	p.StatusBitMask.FragIndex = (data[0] >> 6) & 0x03

	return nil
}

// FragSessionDeleteReqPayload implements the FragSessionDeleteReq paylaod.
type FragSessionDeleteReqPayload struct {
	Param FragSessionDeleteReqPayloadParam
}

// FragSessionDeleteReqPayloadParam implements the FragSessionDeleteReq payload Param field.
type FragSessionDeleteReqPayloadParam struct {
	FragIndex uint8
}

// Size returns the payload size in bytes.
func (p FragSessionDeleteReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionDeleteReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.Param.FragIndex & 0x03
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionDeleteReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d byte is expected", p.Size())
	}
	p.Param.FragIndex = data[0] & 0x3
	return nil
}

// FragSessionDeleteAnsPayload implements the FragSessionDeleteAns payload.
type FragSessionDeleteAnsPayload struct {
	Status FragSessionDeleteAnsPayloadStatus
}

// FragSessionDeleteAnsPayloadStatus implements the FragSessionDeleteAns payload Status field.
type FragSessionDeleteAnsPayloadStatus struct {
	FragIndex           uint8
	SessionDoesNotExist bool
}

// Size returns the size of the payload in bytes.
func (p FragSessionDeleteAnsPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionDeleteAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.Status.FragIndex & 0x03
	if p.Status.SessionDoesNotExist {
		b[0] |= 0x04
	}
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionDeleteAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d byte is expected", p.Size())
	}

	p.Status.FragIndex = data[0] & 0x03
	p.Status.SessionDoesNotExist = data[0]&0x04 != 0

	return nil
}

// DataFragmentPayload implements the DataFragment payload.
type DataFragmentPayload struct {
	IndexAndN DataFragmentPayloadIndexAndN
	Payload   []byte
}

// DataFragmentPayloadIndexAndN implements the DataFragment payload IndexAndN field.
type DataFragmentPayloadIndexAndN struct {
	FragIndex uint8
	N         uint16
}

// Size returns the payload size in bytes.
func (p DataFragmentPayload) Size() int {
	return 2 + len(p.Payload)
}

// MarshalBinary encodes the given payload to a slice of bytes.
func (p DataFragmentPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	binary.LittleEndian.PutUint16(b[0:2], p.IndexAndN.N&0x3fff)
	b[1] |= (p.IndexAndN.FragIndex & 0x03) << 6
	copy(b[2:], p.Payload)

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *DataFragmentPayload) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.New("lorawan/applayer/fragmentation: 2 bytes are expected")
	}

	p.IndexAndN.N = binary.LittleEndian.Uint16(data[0:2]) & 0x3fff // filter out the FragIndex
	p.IndexAndN.FragIndex = data[1] >> 6
	p.Payload = make([]byte, len(data[2:]))
	copy(p.Payload, data[2:])

	return nil
}

// FragSessionStatusReqPayload implements the FragSessionStatusReq payload.
type FragSessionStatusReqPayload struct {
	FragStatusReqParam FragSessionStatusReqPayloadFragStatusReqParam
}

// FragSessionStatusReqPayloadFragStatusReqParam implements the FragSessionStatusReq payload FragStatusReqParam field.
type FragSessionStatusReqPayloadFragStatusReqParam struct {
	FragIndex    uint8
	Participants bool
}

// Size returns the payload size in number of bytes.
func (p FragSessionStatusReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionStatusReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	if p.FragStatusReqParam.Participants {
		b[0] |= 0x01
	}

	b[0] |= (p.FragStatusReqParam.FragIndex & 0x03) << 1

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionStatusReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d byte is expected", p.Size())
	}

	p.FragStatusReqParam.Participants = data[0]&0x01 != 0
	p.FragStatusReqParam.FragIndex = (data[0] >> 1) & 0x03

	return nil
}

// FragSessionStatusAnsPayload implements the FragSessionStatusAns payload.
type FragSessionStatusAnsPayload struct {
	ReceivedAndIndex FragSessionStatusAnsPayloadReceivedAndIndex
	MissingFrag      uint8
	Status           FragSessionStatusAnsPayloadStatus
}

// FragSessionStatusAnsPayloadReceivedAndIndex implements the FragSessionStatusAns payload ReceivedAndIndex field.
type FragSessionStatusAnsPayloadReceivedAndIndex struct {
	FragIndex      uint8
	NbFragReceived uint16
}

// FragSessionStatusAnsPayloadStatus implements the FragSessionStatusAns payload Status field.
type FragSessionStatusAnsPayloadStatus struct {
	NotEnoughMatrixMemory bool
}

// Size returns the payload size in number of bytes.
func (p FragSessionStatusAnsPayload) Size() int {
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p FragSessionStatusAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	binary.LittleEndian.PutUint16(b[0:2], p.ReceivedAndIndex.NbFragReceived&0x3fff)
	b[1] |= (p.ReceivedAndIndex.FragIndex & 0x03) << 6

	b[2] = p.MissingFrag
	if p.Status.NotEnoughMatrixMemory {
		b[3] |= 0x01
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *FragSessionStatusAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/fragmentation: %d bytes are expected", p.Size())
	}

	p.ReceivedAndIndex.NbFragReceived = binary.LittleEndian.Uint16(data[0:2]) & 0x3fff // filter out FragIndex
	p.ReceivedAndIndex.FragIndex = data[1] >> 6

	p.MissingFrag = data[2]
	p.Status.NotEnoughMatrixMemory = data[3]&0x01 != 0

	return nil
}
