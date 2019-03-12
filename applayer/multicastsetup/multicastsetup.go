//go:generate stringer -type=CID

// Package multicastsetup implements the Remote Multicast Setup v1.0.0 over LoRaWAN.
package multicastsetup

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/brocaar/lorawan"
)

// CID defines the command identifier.
type CID byte

// DefaultFPort defines the default fPort value for Remote Multicast Setup.
const DefaultFPort uint8 = 200

// Available command identifiers.
const (
	PackageVersionReq  CID = 0x00
	PackageVersionAns  CID = 0x00
	McGroupStatusReq   CID = 0x01
	McGroupStatusAns   CID = 0x01
	McGroupSetupReq    CID = 0x02
	McGroupSetupAns    CID = 0x02
	McGroupDeleteReq   CID = 0x03
	McGroupDeleteAns   CID = 0x03
	McClassCSessionReq CID = 0x04
	McClassCSessionAns CID = 0x04
	McClassBSessionReq CID = 0x05
	McClassBSessionAns CID = 0x05
)

// Errors
var (
	ErrNoPayloadForCID = errors.New("lorawan/applayer/multicastsetup: no payload for given CID")
)

// map[uplink]...
var commandPayloadRegistry = map[bool]map[CID]func() CommandPayload{
	true: map[CID]func() CommandPayload{
		PackageVersionAns:  func() CommandPayload { return &PackageVersionAnsPayload{} },
		McGroupStatusAns:   func() CommandPayload { return &McGroupStatusAnsPayload{} },
		McGroupSetupAns:    func() CommandPayload { return &McGroupSetupAnsPayload{} },
		McGroupDeleteAns:   func() CommandPayload { return &McGroupDeleteAnsPayload{} },
		McClassCSessionAns: func() CommandPayload { return &McClassCSessionAnsPayload{} },
		McClassBSessionAns: func() CommandPayload { return &McClassBSessionAnsPayload{} },
	},
	false: map[CID]func() CommandPayload{
		McGroupStatusReq:   func() CommandPayload { return &McGroupStatusReqPayload{} },
		McGroupSetupReq:    func() CommandPayload { return &McGroupSetupReqPayload{} },
		McGroupDeleteReq:   func() CommandPayload { return &McGroupDeleteReqPayload{} },
		McClassCSessionReq: func() CommandPayload { return &McClassCSessionReqPayload{} },
		McClassBSessionReq: func() CommandPayload { return &McClassBSessionReqPayload{} },
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
		return errors.New("lorawan/applayer/multicastsetup: at least 1 byte is expected")
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
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	p.PackageIdentifier = data[0]
	p.PackageVersion = data[1]
	return nil
}

// McGroupStatusReqPayload implements the McGroupStatusReq payload.
type McGroupStatusReqPayload struct {
	CmdMask McGroupStatusReqPayloadCmdMask
}

// McGroupStatusReqPayloadCmdMask implements the McGroupStatusReq payload CmdMask field.
type McGroupStatusReqPayloadCmdMask struct {
	RegGroupMask [4]bool
}

// Size returns the payload size in number of bytes.
func (p McGroupStatusReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupStatusReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	for i, mask := range p.CmdMask.RegGroupMask {
		if mask {
			b[0] |= (1 << uint8(i))
		}
	}
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupStatusReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	for i := range p.CmdMask.RegGroupMask {
		p.CmdMask.RegGroupMask[i] = data[0]&(1<<uint8(i)) != 0
	}

	return nil
}

// McGroupStatusAnsPayload implements the McGroupStatusAns payload.
type McGroupStatusAnsPayload struct {
	Status McGroupStatusAnsPayloadStatus
	Items  []McGroupStatusAnsPayloadItem
}

// McGroupStatusAnsPayloadStatus implements the McGroupStatusAns payload Status field.
type McGroupStatusAnsPayloadStatus struct {
	NbTotalGroups uint8
	AnsGroupMask  [4]bool
}

// McGroupStatusAnsPayloadItem implements an McGroupID + MacAddr item.
type McGroupStatusAnsPayloadItem struct {
	McGroupID uint8
	McAddr    lorawan.DevAddr
}

// Size returns the payload size in number of bytes.
func (p McGroupStatusAnsPayload) Size() int {
	var ansGroupMaskCount int
	for _, mask := range p.Status.AnsGroupMask {
		if mask {
			ansGroupMaskCount++
		}
	}

	return 1 + (5 * ansGroupMaskCount)
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupStatusAnsPayload) MarshalBinary() ([]byte, error) {
	if len(p.Items) > 4 {
		return nil, errors.New("lorawan/applayer/multicastsetup: max number of items is 4")
	}

	var ansGroupMaskCount int
	b := make([]byte, 1+(5*len(p.Items)))

	for i, mask := range p.Status.AnsGroupMask {
		if mask {
			b[0] |= (1 << uint8(i))
			ansGroupMaskCount++
		}
	}

	if ansGroupMaskCount != len(p.Items) {
		return nil, errors.New("lorawan/applayer/multicastsetup: number of items does not match AnsGroupMatch")
	}

	b[0] |= (p.Status.NbTotalGroups & 0x07) << 4 // 3 bits

	for i := range p.Items {
		offset := 1 + (i * 5)
		b[offset] = p.Items[i].McGroupID & 0x03 // first two bits

		bb, err := p.Items[i].McAddr.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset+1:offset+5], bb)
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupStatusAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("lorawan/applayer/multicastsetup: at least 1 byte is expected")
	}

	var ansGroupMaskCount int
	for i := range p.Status.AnsGroupMask {
		if data[0]&(1<<uint8(i)) != 0 {
			p.Status.AnsGroupMask[i] = true
			ansGroupMaskCount++
		}
	}

	p.Status.NbTotalGroups = (data[0] & 0x70) >> 4

	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	for i := 0; i < ansGroupMaskCount; i++ {
		offset := 1 + (i * 5)
		item := McGroupStatusAnsPayloadItem{
			McGroupID: data[offset] & 0x03, // first two bits
		}

		if err := item.McAddr.UnmarshalBinary(data[offset+1 : offset+5]); err != nil {
			return err
		}

		p.Items = append(p.Items, item)
	}

	return nil
}

// McGroupSetupReqPayload implements the McGroupSetupReq payload.
type McGroupSetupReqPayload struct {
	McGroupIDHeader McGroupSetupReqPayloadMcGroupIDHeader
	McAddr          lorawan.DevAddr
	McKeyEncrypted  [16]byte
	MinMcFCnt       uint32
	MaxMcFCnt       uint32
}

// McGroupSetupReqPayloadMcGroupIDHeader implements the McGroupSetupReq payload McGroupIDHeader field.
type McGroupSetupReqPayloadMcGroupIDHeader struct {
	McGroupID uint8
}

// Size returns the payload size in number of bytes.
func (p McGroupSetupReqPayload) Size() int {
	return 29
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupSetupReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	// McGroupIDHeader
	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first 2 bits

	// McAddr
	bb, err := p.McAddr.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(b[1:5], bb)

	// The McKeyEncrypted is copied as-is.
	copy(b[5:21], p.McKeyEncrypted[:])

	// MinMcFCnt
	binary.LittleEndian.PutUint32(b[21:25], p.MinMcFCnt)

	// MaxMcFCnt
	binary.LittleEndian.PutUint32(b[25:29], p.MaxMcFCnt)

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	// McAddr
	p.McGroupIDHeader.McGroupID = data[0] & 0x03 // 2 bits

	// McAddr
	if err := p.McAddr.UnmarshalBinary(data[1:5]); err != nil {
		return err
	}

	// The McKeyEncrypted is copied as-is.
	copy(p.McKeyEncrypted[:], data[5:21])

	// MinMcFCnt
	p.MinMcFCnt = binary.LittleEndian.Uint32(data[21:25])

	// MaxMcFCnt
	p.MaxMcFCnt = binary.LittleEndian.Uint32(data[25:29])

	return nil
}

// McGroupSetupAnsPayload implements the McGroupSetupAns payload.
type McGroupSetupAnsPayload struct {
	McGroupIDHeader McGroupSetupAnsPayloadMcGroupIDHeader
}

// McGroupSetupAnsPayloadMcGroupIDHeader implements the McGroupSetupAns payload GroupIDHeader field.
type McGroupSetupAnsPayloadMcGroupIDHeader struct {
	IDError   bool
	McGroupID uint8
}

// Size returns the payload size in number of bytes.
func (p McGroupSetupAnsPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupSetupAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first 2 bits
	if p.McGroupIDHeader.IDError {
		b[0] |= 0x04
	}
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupSetupAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	p.McGroupIDHeader.McGroupID = data[0] & 0x03 // first 2 bits
	p.McGroupIDHeader.IDError = data[0]&0x04 != 0

	return nil
}

// McGroupDeleteReqPayload implements the McGroupDeleteReq payload.
type McGroupDeleteReqPayload struct {
	McGroupIDHeader McGroupDeleteReqPayloadMcGroupIDHeader
}

// McGroupDeleteReqPayloadMcGroupIDHeader implements the McGroupDeleteReq payload McGroupIDHeader field.
type McGroupDeleteReqPayloadMcGroupIDHeader struct {
	McGroupID uint8
}

// Size returns the payload size in number of bytes.
func (p McGroupDeleteReqPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupDeleteReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())
	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first 2 bits
	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupDeleteReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}
	p.McGroupIDHeader.McGroupID = data[0] & 0x03 // first 2 bits
	return nil
}

// McGroupDeleteAnsPayload implements the McGroupDeleteAns payload.
type McGroupDeleteAnsPayload struct {
	McGroupIDHeader McGroupDeleteAnsPayloadMcGroupIDHeader
}

// McGroupDeleteAnsPayloadMcGroupIDHeader implements the McGroupDeleteAns payload McGroupIDHeader field.
type McGroupDeleteAnsPayloadMcGroupIDHeader struct {
	McGroupUndefined bool
	McGroupID        uint8
}

// Size returns the payload size in number of bytes.
func (p McGroupDeleteAnsPayload) Size() int {
	return 1
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McGroupDeleteAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first two bits
	if p.McGroupIDHeader.McGroupUndefined {
		b[0] |= 0x04
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McGroupDeleteAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	p.McGroupIDHeader.McGroupID = data[0] & 0x03
	p.McGroupIDHeader.McGroupUndefined = data[0]&0x04 != 0

	return nil
}

// McClassCSessionReqPayload implements the McClassCSessionReq payload.
type McClassCSessionReqPayload struct {
	McGroupIDHeader McClassCSessionReqPayloadMcGroupIDHeader
	SessionTime     uint32
	SessionTimeOut  McClassCSessionReqPayloadSessionTimeOut
	DLFrequency     uint32 // the frequency in Hz!
	DR              uint8
}

// McClassCSessionReqPayloadMcGroupIDHeader implements the McClassCSessionReq payload McGroupIDHeader field.
type McClassCSessionReqPayloadMcGroupIDHeader struct {
	McGroupID uint8
}

// McClassCSessionReqPayloadSessionTimeOut implements the McClassCSessionReq payload SessionTimeOut field.
type McClassCSessionReqPayloadSessionTimeOut struct {
	TimeOut uint8 // the actual value in seconds is 2^TimeOut
}

// Size returns the payload size in number of bytes.
func (p *McClassCSessionReqPayload) Size() int {
	return 10
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McClassCSessionReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	// McGroupIDHeader
	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first 2 bits

	// SessionTime
	binary.LittleEndian.PutUint32(b[1:5], p.SessionTime)

	// SessionTimeOut
	b[5] = p.SessionTimeOut.TimeOut & 0x0f // first 4 bits

	// DLFrequency
	if p.DLFrequency%100 != 0 {
		return nil, errors.New("lorawan/applayer/multicastsetup: DLFrequency must be a multiple of 100")
	}
	dlFreqB := make([]byte, 4)
	binary.LittleEndian.PutUint32(dlFreqB, p.DLFrequency/100)
	copy(b[6:9], dlFreqB[:3])

	// DR
	b[9] = p.DR

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McClassCSessionReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	// McGroupIDHeader
	p.McGroupIDHeader.McGroupID = data[0] & 0x03

	// SessionTime
	p.SessionTime = binary.LittleEndian.Uint32(data[1:5])

	// SessionTimeOut
	p.SessionTimeOut.TimeOut = data[5] & 0x0f

	// DLFrequency
	dlFreqB := make([]byte, 4)
	copy(dlFreqB, data[6:9])
	p.DLFrequency = binary.LittleEndian.Uint32(dlFreqB) * 100

	// DR
	p.DR = data[9]

	return nil
}

// McClassCSessionAnsPayload implements the McClassCSessionAns payload.
type McClassCSessionAnsPayload struct {
	StatusAndMcGroupID McClassCSessionAnsPayloadStatusAndMcGroupID
	TimeToStart        *uint32
}

// McClassCSessionAnsPayloadStatusAndMcGroupID implements the McClassCSessionAns payload StatusAndMcGroupID field.
type McClassCSessionAnsPayloadStatusAndMcGroupID struct {
	McGroupUndefined bool
	FreqError        bool
	DRError          bool
	McGroupID        uint8
}

func (p McClassCSessionAnsPayloadStatusAndMcGroupID) hasError() bool {
	if p.McGroupUndefined || p.FreqError || p.DRError {
		return true
	}
	return false
}

// Size returns the payload size in number of bytes.
func (p McClassCSessionAnsPayload) Size() int {
	if p.StatusAndMcGroupID.hasError() {
		return 1
	}
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McClassCSessionAnsPayload) MarshalBinary() ([]byte, error) {
	if p.StatusAndMcGroupID.hasError() && p.TimeToStart != nil {
		return nil, errors.New("lorawan/applayer/multicastsetup: TimeToStart must be nil when StatusAndMcGroupID contains an error")
	}

	if !p.StatusAndMcGroupID.hasError() && p.TimeToStart == nil {
		return nil, errors.New("lorawan/applayer/multicastsetup: TimeToStart must not be nil")
	}

	b := make([]byte, 1)

	// StatusAndMcGroupID
	b[0] = p.StatusAndMcGroupID.McGroupID & 0x03 // first 2 bits
	if p.StatusAndMcGroupID.DRError {
		b[0] |= 0x04
	}
	if p.StatusAndMcGroupID.FreqError {
		b[0] |= 0x08
	}
	if p.StatusAndMcGroupID.McGroupUndefined {
		b[0] |= 0x10
	}

	// TimeToStart
	if !p.StatusAndMcGroupID.hasError() {
		ttsB := make([]byte, 4)
		binary.LittleEndian.PutUint32(ttsB, *p.TimeToStart)
		b = append(b, ttsB[:3]...)
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McClassCSessionAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("lorawan/applayer/multicastsetup: at least 1 byte is expected")
	}

	// StatusAndMCGroupID
	p.StatusAndMcGroupID.McGroupID = data[0] & 0x03
	p.StatusAndMcGroupID.DRError = data[0]&0x04 != 0
	p.StatusAndMcGroupID.FreqError = data[0]&0x08 != 0
	p.StatusAndMcGroupID.McGroupUndefined = data[0]&0x10 != 0

	if !p.StatusAndMcGroupID.hasError() {
		if len(data) < p.Size() {
			return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
		}

		ttsB := make([]byte, 4)
		copy(ttsB, data[1:4])
		tts := binary.LittleEndian.Uint32(ttsB)
		p.TimeToStart = &tts
	}

	return nil
}

// McClassBSessionReqPayload implements the McClassBSessionReq payload.
type McClassBSessionReqPayload struct {
	McGroupIDHeader    McClassBSessionReqPayloadMcGroupIDHeader
	SessionTime        uint32
	TimeOutPeriodicity McClassBSessionReqPayloadTimeOutPeriodicity
	DLFrequency        uint32
	DR                 uint8
}

// McClassBSessionReqPayloadMcGroupIDHeader implements the McClassBSessionReq payload McGroupIDHeader field.
type McClassBSessionReqPayloadMcGroupIDHeader struct {
	McGroupID uint8
}

// McClassBSessionReqPayloadTimeOutPeriodicity implements the McClassBSessionReq payload TimeOutPeriodicity field.
type McClassBSessionReqPayloadTimeOutPeriodicity struct {
	Periodicity uint8
	TimeOut     uint8
}

// Size returns the payload size in number of bytes.
func (p McClassBSessionReqPayload) Size() int {
	return 10
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McClassBSessionReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.Size())

	// McGroupIDHeader
	b[0] = p.McGroupIDHeader.McGroupID & 0x03 // first 2 bits

	// SessionTime
	binary.LittleEndian.PutUint32(b[1:5], p.SessionTime)

	// TimeOutPeriodicity
	b[5] = p.TimeOutPeriodicity.TimeOut & 0x1f // first 4 bits
	b[5] |= (p.TimeOutPeriodicity.Periodicity & 0x17) << 4

	// DLFrequency
	if p.DLFrequency%100 != 0 {
		return nil, errors.New("lorawan/applayer/multicastsetup: DLFrequency must be a multiple of 100")
	}
	dlFreqB := make([]byte, 4)
	binary.LittleEndian.PutUint32(dlFreqB, p.DLFrequency/100)
	copy(b[6:9], dlFreqB[:3])

	// DR
	b[9] = p.DR

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McClassBSessionReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) < p.Size() {
		return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
	}

	// McGroupIDHeader
	p.McGroupIDHeader.McGroupID = data[0] & 0x03

	// SessionTime
	p.SessionTime = binary.LittleEndian.Uint32(data[1:5])

	// TimeOutPeriodicity
	p.TimeOutPeriodicity.TimeOut = data[5] & 0x1f
	p.TimeOutPeriodicity.Periodicity = (data[5] >> 4) & 0x17

	// DLFrequency
	dlFreqB := make([]byte, 4)
	copy(dlFreqB, data[6:9])
	p.DLFrequency = binary.LittleEndian.Uint32(dlFreqB) * 100

	// DR
	p.DR = data[9]

	return nil
}

// McClassBSessionAnsPayload implements the McClassBSessionAns payload.
type McClassBSessionAnsPayload struct {
	StatusAndMcGroupID McClassBSessionAnsPayloadStatusAndMcGroupID
	TimeToStart        *uint32
}

// McClassBSessionAnsPayloadStatusAndMcGroupID implements the McClassBSessionAns payload StatusAndMcGroupID field.
type McClassBSessionAnsPayloadStatusAndMcGroupID struct {
	McGroupUndefined bool
	FreqError        bool
	DRError          bool
	McGroupID        uint8
}

func (p McClassBSessionAnsPayloadStatusAndMcGroupID) hasError() bool {
	if p.McGroupUndefined || p.FreqError || p.DRError {
		return true
	}
	return false
}

// Size returns the payload size in number of bytes.
func (p McClassBSessionAnsPayload) Size() int {
	if p.StatusAndMcGroupID.hasError() {
		return 1
	}
	return 4
}

// MarshalBinary encodes the payload to a slice of bytes.
func (p McClassBSessionAnsPayload) MarshalBinary() ([]byte, error) {
	if p.StatusAndMcGroupID.hasError() && p.TimeToStart != nil {
		return nil, errors.New("lorawan/applayer/multicastsetup: TimeToStart must be nil when StatusAndMcGroupID contains an error")
	}

	if !p.StatusAndMcGroupID.hasError() && p.TimeToStart == nil {
		return nil, errors.New("lorawan/applayer/multicastsetup: TimeToStart must not be nil")
	}

	b := make([]byte, 1)

	// StatusAndMcGroupID
	b[0] = p.StatusAndMcGroupID.McGroupID & 0x03 // first 2 bits
	if p.StatusAndMcGroupID.DRError {
		b[0] |= 0x04
	}
	if p.StatusAndMcGroupID.FreqError {
		b[0] |= 0x08
	}
	if p.StatusAndMcGroupID.McGroupUndefined {
		b[0] |= 0x10
	}

	if !p.StatusAndMcGroupID.hasError() {
		ttsB := make([]byte, 4)
		binary.LittleEndian.PutUint32(ttsB, *p.TimeToStart)
		b = append(b, ttsB[:3]...)
	}

	return b, nil
}

// UnmarshalBinary decodes the payload from a slice of bytes.
func (p *McClassBSessionAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("lorawan/applayer/multicastsetup: at least 1 byte is expected")
	}

	// StatusAndMcGroupID
	p.StatusAndMcGroupID.McGroupID = data[0] & 0x03
	p.StatusAndMcGroupID.DRError = data[0]&0x04 != 0
	p.StatusAndMcGroupID.FreqError = data[0]&0x08 != 0
	p.StatusAndMcGroupID.McGroupUndefined = data[0]&0x10 != 0

	if !p.StatusAndMcGroupID.hasError() {
		if len(data) < p.Size() {
			return fmt.Errorf("lorawan/applayer/multicastsetup: %d bytes are expected", p.Size())
		}

		ttsB := make([]byte, 4)
		copy(ttsB, data[1:4])
		tts := binary.LittleEndian.Uint32(ttsB)
		p.TimeToStart = &tts
	}

	return nil
}
