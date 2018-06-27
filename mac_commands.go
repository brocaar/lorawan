//go:generate stringer -type=CID

package lorawan

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// macPayloadMutex is used when registering proprietary MAC command payloads to
// the macPayloadRegistry.
var macPayloadMutex sync.RWMutex

// CID defines the MAC command identifier.
type CID byte

// MarshalText implements encoding.TextMarshaler.
func (c CID) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

// MAC commands as specified by the LoRaWAN R1.0 specs. Note that each *Req / *Ans
// has the same value. Based on the fact if a message is uplink or downlink
// you should use on or the other.
const (
	ResetInd            CID = 0x01
	ResetConf           CID = 0x01
	LinkCheckReq        CID = 0x02
	LinkCheckAns        CID = 0x02
	LinkADRReq          CID = 0x03
	LinkADRAns          CID = 0x03
	DutyCycleReq        CID = 0x04
	DutyCycleAns        CID = 0x04
	RXParamSetupReq     CID = 0x05
	RXParamSetupAns     CID = 0x05
	DevStatusReq        CID = 0x06
	DevStatusAns        CID = 0x06
	NewChannelReq       CID = 0x07
	NewChannelAns       CID = 0x07
	RXTimingSetupReq    CID = 0x08
	RXTimingSetupAns    CID = 0x08
	TXParamSetupReq     CID = 0x09
	TXParamSetupAns     CID = 0x09
	DLChannelReq        CID = 0x0A
	DLChannelAns        CID = 0x0A
	RekeyInd            CID = 0x0B
	RekeyConf           CID = 0x0B
	ADRParamSetupReq    CID = 0x0C
	ADRParamSetupAns    CID = 0x0C
	DeviceTimeReq       CID = 0x0D
	DeviceTimeAns       CID = 0x0D
	ForceRejoinReq      CID = 0x0E
	RejoinParamSetupReq CID = 0x0F
	RejoinParamSetupAns CID = 0x0F
	PingSlotInfoReq     CID = 0x10
	PingSlotInfoAns     CID = 0x10
	PingSlotChannelReq  CID = 0x11
	PingSlotChannelAns  CID = 0x11
	// 0x12 has been deprecated in 1.1
	BeaconFreqReq CID = 0x13
	BeaconFreqAns CID = 0x13
	// 0x80 to 0xFF reserved for proprietary network command extensions
)

// macPayloadInfo contains the info about a MAC payload
type macPayloadInfo struct {
	size    int
	payload func() MACCommandPayload
}

// macPayloadRegistry contains the info for uplink and downlink MAC payloads
// in the format map[uplink]map[CID].
// Note that MAC command that do not have a payload are not included in this
// list.
var macPayloadRegistry = map[bool]map[CID]macPayloadInfo{
	false: map[CID]macPayloadInfo{
		ResetConf:           {1, func() MACCommandPayload { return &ResetConfPayload{} }},
		LinkCheckAns:        {2, func() MACCommandPayload { return &LinkCheckAnsPayload{} }},
		LinkADRReq:          {4, func() MACCommandPayload { return &LinkADRReqPayload{} }},
		DutyCycleReq:        {1, func() MACCommandPayload { return &DutyCycleReqPayload{} }},
		RXParamSetupReq:     {4, func() MACCommandPayload { return &RXParamSetupReqPayload{} }},
		NewChannelReq:       {5, func() MACCommandPayload { return &NewChannelReqPayload{} }},
		RXTimingSetupReq:    {1, func() MACCommandPayload { return &RXTimingSetupReqPayload{} }},
		TXParamSetupReq:     {1, func() MACCommandPayload { return &TXParamSetupReqPayload{} }},
		DLChannelReq:        {4, func() MACCommandPayload { return &DLChannelReqPayload{} }},
		BeaconFreqReq:       {3, func() MACCommandPayload { return &BeaconFreqReqPayload{} }},
		PingSlotChannelReq:  {4, func() MACCommandPayload { return &PingSlotChannelReqPayload{} }},
		DeviceTimeAns:       {5, func() MACCommandPayload { return &DeviceTimeAnsPayload{} }},
		RekeyConf:           {1, func() MACCommandPayload { return &RekeyConfPayload{} }},
		ADRParamSetupReq:    {1, func() MACCommandPayload { return &ADRParamSetupReqPayload{} }},
		ForceRejoinReq:      {2, func() MACCommandPayload { return &ForceRejoinReqPayload{} }},
		RejoinParamSetupReq: {1, func() MACCommandPayload { return &RejoinParamSetupReqPayload{} }},
	},
	true: map[CID]macPayloadInfo{
		ResetInd:            {1, func() MACCommandPayload { return &ResetIndPayload{} }},
		LinkADRAns:          {1, func() MACCommandPayload { return &LinkADRAnsPayload{} }},
		RXParamSetupAns:     {1, func() MACCommandPayload { return &RXParamSetupAnsPayload{} }},
		DevStatusAns:        {2, func() MACCommandPayload { return &DevStatusAnsPayload{} }},
		NewChannelAns:       {1, func() MACCommandPayload { return &NewChannelAnsPayload{} }},
		DLChannelAns:        {1, func() MACCommandPayload { return &DLChannelAnsPayload{} }},
		PingSlotInfoReq:     {1, func() MACCommandPayload { return &PingSlotInfoReqPayload{} }},
		BeaconFreqAns:       {1, func() MACCommandPayload { return &BeaconFreqAnsPayload{} }},
		PingSlotChannelAns:  {1, func() MACCommandPayload { return &PingSlotChannelAnsPayload{} }},
		RekeyInd:            {1, func() MACCommandPayload { return &RekeyIndPayload{} }},
		RejoinParamSetupAns: {1, func() MACCommandPayload { return &RejoinParamSetupAnsPayload{} }},
	},
}

// DwellTime defines the dwell time type.
type DwellTime int

// Possible dwell time options.
const (
	DwellTimeNoLimit DwellTime = iota
	DwellTime400ms
)

// GetMACPayloadAndSize returns a new MACCommandPayload instance and it's size.
func GetMACPayloadAndSize(uplink bool, c CID) (MACCommandPayload, int, error) {
	macPayloadMutex.RLock()
	defer macPayloadMutex.RUnlock()

	v, ok := macPayloadRegistry[uplink][c]
	if !ok {
		return nil, 0, fmt.Errorf("lorawan: payload unknown for uplink=%v and CID=%v", uplink, c)
	}

	return v.payload(), v.size, nil
}

// RegisterProprietaryMACCommand registers a proprietary MAC command. Note
// that there is no need to call this when the size of the payload is > 0 bytes.
func RegisterProprietaryMACCommand(uplink bool, cid CID, payloadSize int) error {
	if !(cid >= 128 && cid <= 255) {
		return fmt.Errorf("lorawan: invalid CID %x", byte(cid))
	}

	if payloadSize == 0 {
		// no need to register the payload size
		return nil
	}

	macPayloadMutex.Lock()
	defer macPayloadMutex.Unlock()

	macPayloadRegistry[uplink][cid] = macPayloadInfo{
		size:    payloadSize,
		payload: func() MACCommandPayload { return &ProprietaryMACCommandPayload{} },
	}

	return nil
}

// MACCommandPayload is the interface that every MACCommand payload
// must implement.
type MACCommandPayload interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}

// MACCommand represents a MAC command with optional payload.
type MACCommand struct {
	CID     CID               `json:"cid"`
	Payload MACCommandPayload `json:"payload"`
}

// MarshalBinary marshals the object in binary form.
func (m MACCommand) MarshalBinary() ([]byte, error) {
	b := []byte{byte(m.CID)}

	if m.Payload != nil {
		p, err := m.Payload.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b = append(b, p...)
	}
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (m *MACCommand) UnmarshalBinary(uplink bool, data []byte) error {
	if len(data) == 0 {
		return errors.New("lorawan: at least 1 byte of data is expected")
	}

	m.CID = CID(data[0])

	if len(data) > 1 {
		p, _, err := GetMACPayloadAndSize(uplink, m.CID)
		if err != nil {
			return err
		}
		m.Payload = p
		if err := m.Payload.UnmarshalBinary(data[1:]); err != nil {
			return err
		}
	}
	return nil
}

// ProprietaryMACCommandPayload represents a proprietary payload.
type ProprietaryMACCommandPayload struct {
	Bytes []byte `json:"bytes"`
}

// MarshalBinary marshals the object into a slice of bytes.
func (p ProprietaryMACCommandPayload) MarshalBinary() ([]byte, error) {
	return p.Bytes, nil
}

// UnmarshalBinary decodes the object from a slice of bytes.
func (p *ProprietaryMACCommandPayload) UnmarshalBinary(data []byte) error {
	p.Bytes = data
	return nil
}

// LinkCheckAnsPayload represents the LinkCheckAns payload.
type LinkCheckAnsPayload struct {
	Margin uint8 `json:"margin"`
	GwCnt  uint8 `json:"gwCnt"`
}

// MarshalBinary marshals the object in binary form.
func (p LinkCheckAnsPayload) MarshalBinary() ([]byte, error) {
	return []byte{byte(p.Margin), byte(p.GwCnt)}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *LinkCheckAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("lorawan: 2 bytes of data are expected")
	}
	p.Margin = uint8(data[0])
	p.GwCnt = uint8(data[1])
	return nil
}

// ChMask encodes the channels usable for uplink access. 0 = channel 1,
// 15 = channel 16.
type ChMask [16]bool

// MarshalBinary marshals the object in binary form.
func (m ChMask) MarshalBinary() ([]byte, error) {
	b := make([]byte, 2)
	for i := uint8(0); i < 16; i++ {
		if m[i] {
			b[i/8] = b[i/8] ^ 1<<(i%8)
		}
	}
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (m *ChMask) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("lorawan: 2 bytes of data are expected")
	}
	for i, b := range data {
		for j := uint8(0); j < 8; j++ {
			if b&(1<<j) > 0 {
				m[uint8(i)*8+j] = true
			}
		}
	}
	return nil
}

// Redundancy represents the redundancy field.
type Redundancy struct {
	ChMaskCntl uint8 `json:"chMaskCntl"`
	NbRep      uint8 `json:"nbRep"`
}

// MarshalBinary marshals the object in binary form.
func (r Redundancy) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	if r.NbRep > 15 {
		return b, errors.New("lorawan: max value of NbRep is 15")
	}
	if r.ChMaskCntl > 7 {
		return b, errors.New("lorawan: max value of ChMaskCntl is 7")
	}
	b[0] = r.NbRep ^ (r.ChMaskCntl << 4)
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (r *Redundancy) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	r.NbRep = data[0] & ((1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0))
	r.ChMaskCntl = (data[0] & ((1 << 6) ^ (1 << 5) ^ (1 << 4))) >> 4
	return nil
}

// LinkADRReqPayload represents the LinkADRReq payload.
type LinkADRReqPayload struct {
	DataRate   uint8      `json:"dataRate"`
	TXPower    uint8      `json:"txPower"`
	ChMask     ChMask     `json:"chMask"`
	Redundancy Redundancy `json:"redundancy"`
}

// MarshalBinary marshals the object in binary form.
func (p LinkADRReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 4)
	if p.DataRate > 15 {
		return b, errors.New("lorawan: the max value of DataRate is 15")
	}
	if p.TXPower > 15 {
		return b, errors.New("lorawan: the max value of TXPower is 15")
	}

	cm, err := p.ChMask.MarshalBinary()
	if err != nil {
		return b, err
	}
	r, err := p.Redundancy.MarshalBinary()
	if err != nil {
		return b, err
	}

	b = append(b, p.TXPower^(p.DataRate<<4))
	b = append(b, cm...)
	b = append(b, r...)

	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *LinkADRReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}
	p.DataRate = (data[0] & ((1 << 7) ^ (1 << 6) ^ (1 << 5) ^ (1 << 4))) >> 4
	p.TXPower = data[0] & ((1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0))

	if err := p.ChMask.UnmarshalBinary(data[1:3]); err != nil {
		return err
	}
	return p.Redundancy.UnmarshalBinary(data[3:4])
}

// LinkADRAnsPayload represents the LinkADRAns payload.
type LinkADRAnsPayload struct {
	ChannelMaskACK bool `json:"channelMaskAck"`
	DataRateACK    bool `json:"dataRateAck"`
	PowerACK       bool `json:"powerAck"`
}

// MarshalBinary marshals the object in binary form.
func (p LinkADRAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.ChannelMaskACK {
		b = b ^ (1 << 0)
	}
	if p.DataRateACK {
		b = b ^ (1 << 1)
	}
	if p.PowerACK {
		b = b ^ (1 << 2)
	}
	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *LinkADRAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	if data[0]&(1<<0) > 0 {
		p.ChannelMaskACK = true
	}
	if data[0]&(1<<1) > 0 {
		p.DataRateACK = true
	}
	if data[0]&(1<<2) > 0 {
		p.PowerACK = true
	}
	return nil
}

// DutyCycleReqPayload represents the DutyCycleReq payload.
type DutyCycleReqPayload struct {
	MaxDCycle uint8 `json:"maxDCycle"`
}

// MarshalBinary marshals the object in binary form.
func (p DutyCycleReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 1)
	if p.MaxDCycle > 15 && p.MaxDCycle < 255 {
		return b, errors.New("lorawan: only a MaxDCycle value of 0 - 15 and 255 is allowed")
	}
	b = append(b, p.MaxDCycle)
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *DutyCycleReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	p.MaxDCycle = data[0]
	return nil
}

// DLSettings represents the DLSettings fields (downlink settings).
type DLSettings struct {
	OptNeg      bool  `json:"optNeg"`
	RX2DataRate uint8 `json:"rx2DataRate"`
	RX1DROffset uint8 `json:"rx1DROffset"`
}

// MarshalBinary marshals the object in binary form.
func (s DLSettings) MarshalBinary() ([]byte, error) {
	if s.RX2DataRate > 15 {
		return nil, errors.New("lorawan: max value of RX2DataRate is 15")
	}
	if s.RX1DROffset > 7 {
		return nil, errors.New("lorawan: max value of RX1DROffset is 7")
	}

	b := s.RX2DataRate
	b |= s.RX1DROffset << 4

	if s.OptNeg {
		b |= 1 << 7
	}

	return []byte{b}, nil
}

// MarshalText implements encoding.TextMarshaler.
func (s DLSettings) MarshalText() ([]byte, error) {
	b, err := s.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return []byte(hex.EncodeToString(b)), nil
}

// UnmarshalBinary decodes the object from binary form.
func (s *DLSettings) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	s.OptNeg = (data[0] & (1 << 7)) != 0
	s.RX2DataRate = data[0] & ((1 << 3) | (1 << 2) | (1 << 1) | 1)
	s.RX1DROffset = (data[0] & ((1 << 6) | (1 << 5) | (1 << 4))) >> 4

	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *DLSettings) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}

	return s.UnmarshalBinary(b)
}

// RXParamSetupReqPayload represents the RXParamSetupReq payload.
type RXParamSetupReqPayload struct {
	Frequency  uint32     `json:"frequency"`
	DLSettings DLSettings `json:"dlSettings"`
}

// MarshalBinary marshals the object in binary form.
func (p RXParamSetupReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)
	if p.Frequency/100 >= 16777216 { // 2^24
		return b, errors.New("lorawan: max value of Frequency is 2^24-1")
	}
	if p.Frequency%100 != 0 {
		return b, errors.New("lorawan: Frequency must be a multiple of 100")
	}
	bytes, err := p.DLSettings.MarshalBinary()
	if err != nil {
		return b, err
	}
	b[0] = bytes[0]

	binary.LittleEndian.PutUint32(b[1:5], p.Frequency/100)
	// we don't return the last octet which is fine since we're only interested
	// in the 24 LSB of Frequency
	return b[0:4], nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *RXParamSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}
	if err := p.DLSettings.UnmarshalBinary(data[0:1]); err != nil {
		return err
	}
	// append one block of empty bits at the end of the slice since the
	// binary to uint32 expects 32 bits.
	b := make([]byte, len(data))
	copy(b, data)
	b = append(b, byte(0))
	p.Frequency = binary.LittleEndian.Uint32(b[1:5]) * 100
	return nil
}

// RXParamSetupAnsPayload represents the RXParamSetupAns payload.
type RXParamSetupAnsPayload struct {
	ChannelACK     bool `json:"channelAck"`
	RX2DataRateACK bool `json:"rx2DataRateAck"`
	RX1DROffsetACK bool `json:"rx1DROffsetAck"`
}

// MarshalBinary marshals the object in binary form.
func (p RXParamSetupAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.ChannelACK {
		b = b ^ (1 << 0)
	}
	if p.RX2DataRateACK {
		b = b ^ (1 << 1)
	}
	if p.RX1DROffsetACK {
		b = b ^ (1 << 2)
	}
	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *RXParamSetupAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	p.ChannelACK = data[0]&(1<<0) > 0
	p.RX2DataRateACK = data[0]&(1<<1) > 0
	p.RX1DROffsetACK = data[0]&(1<<2) > 0
	return nil
}

// DevStatusAnsPayload represents the DevStatusAns payload.
type DevStatusAnsPayload struct {
	Battery uint8 `json:"battery"`
	Margin  int8  `json:"margin"`
}

// MarshalBinary marshals the object in binary form.
func (p DevStatusAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 2)
	if p.Margin < -32 {
		return b, errors.New("lorawan: min value of Margin is -32")
	}
	if p.Margin > 31 {
		return b, errors.New("lorawan: max value of Margin is 31")
	}

	b = append(b, p.Battery)
	if p.Margin < 0 {
		b = append(b, uint8(64+p.Margin))
	} else {
		b = append(b, uint8(p.Margin))
	}
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *DevStatusAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("lorawan: 2 bytes of data are expected")
	}
	p.Battery = data[0]
	if data[1] > 31 {
		p.Margin = int8(data[1]) - 64
	} else {
		p.Margin = int8(data[1])
	}
	return nil
}

// NewChannelReqPayload represents the NewChannelReq payload.
type NewChannelReqPayload struct {
	ChIndex uint8  `json:"chIndex"`
	Freq    uint32 `json:"freq"`
	MaxDR   uint8  `json:"maxDR"`
	MinDR   uint8  `json:"minDR"`
}

// MarshalBinary marshals the object in binary form.
func (p NewChannelReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)
	if p.Freq/100 >= 16777216 { // 2^24
		return b, errors.New("lorawan: max value of Freq is 2^24 - 1")
	}
	if p.Freq%100 != 0 {
		return b, errors.New("lorawan: Freq must be a multiple of 100")
	}
	if p.MaxDR > 15 {
		return b, errors.New("lorawan: max value of MaxDR is 15")
	}
	if p.MinDR > 15 {
		return b, errors.New("lorawan: max value of MinDR is 15")
	}

	// we're borrowing the last byte b[4] because PutUint32 needs 4 bytes,
	// the last byte b[4] will be set to 0 because max Freq = 2^24 - 1
	binary.LittleEndian.PutUint32(b[1:5], p.Freq/100)
	b[0] = p.ChIndex
	b[4] = p.MinDR ^ (p.MaxDR << 4)

	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *NewChannelReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("lorawan: 5 bytes of data are expected")
	}
	p.ChIndex = data[0]
	p.MinDR = data[4] & ((1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0))
	p.MaxDR = (data[4] & ((1 << 7) ^ (1 << 6) ^ (1 << 5) ^ (1 << 4))) >> 4

	b := make([]byte, len(data))
	copy(b, data)
	b[4] = byte(0)
	p.Freq = binary.LittleEndian.Uint32(b[1:5]) * 100
	return nil
}

// NewChannelAnsPayload represents the NewChannelAns payload.
type NewChannelAnsPayload struct {
	ChannelFrequencyOK bool `json:"channelFrequencyOK"`
	DataRateRangeOK    bool `json:"dataRateRangeOK"`
}

// MarshalBinary marshals the object in binary form.
func (p NewChannelAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.ChannelFrequencyOK {
		b = (1 << 0)
	}
	if p.DataRateRangeOK {
		b = b ^ (1 << 1)
	}
	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *NewChannelAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	p.ChannelFrequencyOK = data[0]&(1<<0) > 0
	p.DataRateRangeOK = data[0]&(1<<1) > 0
	return nil
}

// RXTimingSetupReqPayload represents the RXTimingSetupReq payload.
type RXTimingSetupReqPayload struct {
	Delay uint8 `json:"delay"` // 0=1s, 1=1s, 2=2s, ... 15=15s
}

// MarshalBinary marshals the object in binary form.
func (p RXTimingSetupReqPayload) MarshalBinary() ([]byte, error) {
	if p.Delay > 15 {
		return []byte{}, errors.New("lorawan: the max value of Delay is 15")
	}
	return []byte{p.Delay}, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *RXTimingSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	p.Delay = data[0]
	return nil
}

// TXParamSetupReqPayload represents the TXParamSetupReq payload.
type TXParamSetupReqPayload struct {
	DownlinkDwelltime DwellTime `json:"downlinkDwellTime"`
	UplinkDwellTime   DwellTime `json:"uplinkDwellTime"`
	MaxEIRP           uint8     `json:"maxEIRP"`
}

// MarshalBinary encodes the object into a bytes.
func (p TXParamSetupReqPayload) MarshalBinary() ([]byte, error) {
	var b uint8
	for i, v := range []uint8{8, 10, 12, 13, 14, 16, 18, 20, 21, 24, 26, 27, 29, 30, 33, 36} {
		if v == p.MaxEIRP {
			b = uint8(i)
		}
	}
	if b == 0 {
		return nil, errors.New("lorawan: invalid MaxEIRP value")
	}

	if p.UplinkDwellTime == DwellTime400ms {
		b = b ^ (1 << 4)
	}
	if p.DownlinkDwelltime == DwellTime400ms {
		b = b ^ (1 << 5)
	}

	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *TXParamSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	if data[0]&(1<<4) > 0 {
		p.UplinkDwellTime = DwellTime400ms
	}
	if data[0]&(1<<5) > 0 {
		p.DownlinkDwelltime = DwellTime400ms
	}
	p.MaxEIRP = []uint8{8, 10, 12, 13, 14, 16, 18, 20, 21, 24, 26, 27, 29, 30, 33, 36}[data[0]&15]

	return nil
}

// DLChannelReqPayload represents the DLChannelReq payload.
type DLChannelReqPayload struct {
	ChIndex uint8  `json:"chIndex"`
	Freq    uint32 `json:"freq"`
}

// MarshalBinary encodes the object into bytes.
func (p DLChannelReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)        // we need one byte more for PutUint32
	if p.Freq/100 >= 16777216 { // 2^24
		return b, errors.New("lorawan: max value of Freq is 2^24 - 1")
	}

	if p.Freq%100 != 0 {
		return b, errors.New("lorawan: Freq must be a multiple of 100")
	}

	b[0] = p.ChIndex
	binary.LittleEndian.PutUint32(b[1:5], p.Freq/100)

	return b[0:4], nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *DLChannelReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}

	p.ChIndex = data[0]
	b := make([]byte, 4)
	copy(b, data[1:])
	p.Freq = binary.LittleEndian.Uint32(b) * 100
	return nil
}

// DLChannelAnsPayload represents the DLChannelAns payload.
type DLChannelAnsPayload struct {
	UplinkFrequencyExists bool `json:"uplinkFrequencyExists"`
	ChannelFrequencyOK    bool `json:"channelFrequencyOK"`
}

// MarshalBinary encodes the object into bytes.
func (p DLChannelAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.ChannelFrequencyOK {
		b = b ^ 1
	}
	if p.UplinkFrequencyExists {
		b = b ^ (1 << 1)
	}
	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *DLChannelAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	p.ChannelFrequencyOK = data[0]&1 > 0
	p.UplinkFrequencyExists = data[0]&(1<<1) > 0
	return nil
}

// PingSlotInfoReqPayload represents the PingSlotInfoReq payload.
type PingSlotInfoReqPayload struct {
	Periodicity uint8 `json:"periodicity"`
}

// MarshalBinary encodes the object into bytes.
func (p PingSlotInfoReqPayload) MarshalBinary() ([]byte, error) {
	if p.Periodicity > 7 {
		return nil, errors.New("lorawan: max value of Periodicity is 7")
	}

	return []byte{byte(p.Periodicity)}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *PingSlotInfoReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	// first 3 bits
	p.Periodicity = data[0] & ((1 << 2) | (1 << 1) | (1 << 0))

	return nil
}

// BeaconFreqReqPayload represents the BeaconFreqReq payload.
type BeaconFreqReqPayload struct {
	Frequency uint32 `json:"frequency"`
}

// MarshalBinary encodes the object into bytes.
func (p BeaconFreqReqPayload) MarshalBinary() ([]byte, error) {
	if p.Frequency/100 >= 16777216 { // 2^24
		return nil, errors.New("lorawan: max value of Frequency is 2^24 - 1")
	}
	if p.Frequency%100 != 0 {
		return nil, errors.New("lorawan: Frequency must be a multiple of 100")
	}

	// we need 4 bytes for PutUint32
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.Frequency/100)

	// only return the first 3 bytes
	return b[0:3], nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *BeaconFreqReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 3 {
		return errors.New("lorawan: 3 bytes of data are expected")
	}

	// we need 4 bytes for Uint32
	b := make([]byte, 4)
	copy(b, data)
	p.Frequency = binary.LittleEndian.Uint32(b) * 100

	return nil
}

// BeaconFreqAnsPayload represents the BeaconFreqAns payload.
type BeaconFreqAnsPayload struct {
	BeaconFrequencyOK bool `json:"beaconFrequencyOK"`
}

// MarshalBinary encodes the object into bytes.
func (p BeaconFreqAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.BeaconFrequencyOK {
		b = (1 << 0)
	}

	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *BeaconFreqAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	p.BeaconFrequencyOK = data[0]&(1<<0) != 0

	return nil
}

// PingSlotChannelReqPayload represents the PingSlotChannelReq payload.
type PingSlotChannelReqPayload struct {
	Frequency uint32 `json:"frequency"`
	DR        uint8  `json:"dr"`
}

// MarshalBinary encodes the object into bytes.
func (p PingSlotChannelReqPayload) MarshalBinary() ([]byte, error) {
	if p.Frequency/100 >= 16777216 { // 2^24
		return nil, errors.New("lorawan: max value of Frequency is 2^24 - 1")
	}
	if p.Frequency%100 != 0 {
		return nil, errors.New("lorawan: Frequency must be a multiple of 100")
	}
	if p.DR >= 16 { // 2^4
		return nil, errors.New("lorawan: max value of DR is 15")
	}

	// allocate one extra byte for PutUint32
	b := make([]byte, 4)

	binary.LittleEndian.PutUint32(b, p.Frequency/100)
	b[3] = byte(p.DR)

	return b, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *PingSlotChannelReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}

	b := make([]byte, 4)
	copy(b, data)
	b[3] = 0

	p.Frequency = binary.LittleEndian.Uint32(b) * 100
	p.DR = data[3] & ((1 << 3) | (1 << 2) | (1 << 1) | (1 << 0))

	return nil
}

// PingSlotChannelAnsPayload represents the PingSlotChannelAns payload.
type PingSlotChannelAnsPayload struct {
	DataRateOK         bool `json:"dataRateOK"`
	ChannelFrequencyOK bool `json:"channelFrequencyOK"`
}

// MarshalBinary encodes the object into bytes.
func (p PingSlotChannelAnsPayload) MarshalBinary() ([]byte, error) {
	var b byte
	if p.ChannelFrequencyOK {
		b = (1 << 0)
	}
	if p.DataRateOK {
		b = b | (1 << 1)
	}

	return []byte{b}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *PingSlotChannelAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	p.ChannelFrequencyOK = data[0]&(1<<0) != 0
	p.DataRateOK = data[0]&(1<<1) != 0

	return nil
}

// DeviceTimeAnsPayload represents the DeviceTimeAns payload.
type DeviceTimeAnsPayload struct {
	TimeSinceGPSEpoch time.Duration `json:"timeSinceGPSEpoch"`
}

// MarshalBinary encodes the object into bytes.
func (p DeviceTimeAnsPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)

	seconds := uint32(p.TimeSinceGPSEpoch / time.Second)
	binary.LittleEndian.PutUint32(b, seconds)

	// time.Second / 256 = 3906250ns
	b[4] = uint8((p.TimeSinceGPSEpoch - (time.Duration(seconds) * time.Second)) / 3906250)

	return b, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *DeviceTimeAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 5 {
		return errors.New("lorawan: 5 bytes of data is expected")
	}

	seconds := binary.LittleEndian.Uint32(data[0:4])
	p.TimeSinceGPSEpoch = time.Second * time.Duration(seconds)
	p.TimeSinceGPSEpoch += time.Duration(data[4]) * 3906250

	return nil
}

// Version defines LoRaWAN version field.
type Version struct {
	Minor uint8 `json:"minor"`
}

// MarshalBinary encodes the object into bytes.
func (v Version) MarshalBinary() ([]byte, error) {
	if v.Minor > 7 {
		return nil, errors.New("lorawan: max value of Minor is 7")
	}
	return []byte{v.Minor}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (v *Version) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	v.Minor = data[0]
	return nil
}

// ResetIndPayload represents the ResetInd payload.
type ResetIndPayload struct {
	DevLoRaWANVersion Version `json:"devLoRaWANVersion"`
}

// MarshalBinary encodes the object into bytes.
func (p ResetIndPayload) MarshalBinary() ([]byte, error) {
	return p.DevLoRaWANVersion.MarshalBinary()
}

// UnmarshalBinary decodes the object from bytes.
func (p *ResetIndPayload) UnmarshalBinary(data []byte) error {
	return p.DevLoRaWANVersion.UnmarshalBinary(data)
}

// ResetConfPayload represents the ResetConf payload.
type ResetConfPayload struct {
	ServLoRaWANVersion Version `json:"servLoRaWANVersion"`
}

// MarshalBinary encodes the object into bytes.
func (p ResetConfPayload) MarshalBinary() ([]byte, error) {
	return p.ServLoRaWANVersion.MarshalBinary()
}

// UnmarshalBinary decodes the object from bytes.
func (p *ResetConfPayload) UnmarshalBinary(data []byte) error {
	return p.ServLoRaWANVersion.UnmarshalBinary(data)
}

// RekeyIndPayload represents the RekeyInd payload.
type RekeyIndPayload struct {
	DevLoRaWANVersion Version `json:"devLoRaWANVersion"`
}

// MarshalBinary encodes the object into bytes.
func (p RekeyIndPayload) MarshalBinary() ([]byte, error) {
	return p.DevLoRaWANVersion.MarshalBinary()
}

// UnmarshalBinary decodes the object from bytes.
func (p *RekeyIndPayload) UnmarshalBinary(data []byte) error {
	return p.DevLoRaWANVersion.UnmarshalBinary(data)
}

// RekeyConfPayload represents the RekeyConf payload.
type RekeyConfPayload struct {
	ServLoRaWANVersion Version `json:"servLoRaWANVersion"`
}

// MarshalBinary encodes the object into bytes.
func (p RekeyConfPayload) MarshalBinary() ([]byte, error) {
	return p.ServLoRaWANVersion.MarshalBinary()
}

// UnmarshalBinary decodes the object from bytes.
func (p *RekeyConfPayload) UnmarshalBinary(data []byte) error {
	return p.ServLoRaWANVersion.UnmarshalBinary(data)
}

// ADRParam defines the ADRParam field.
type ADRParam struct {
	LimitExp uint8 `json:"limitExp"`
	DelayExp uint8 `json:"delayExp"`
}

// MarshalBinary encodes the object into bytes.
func (p ADRParam) MarshalBinary() ([]byte, error) {
	if p.LimitExp > 15 {
		return nil, errors.New("lorawan: max value of LimitExp is 15")
	}
	if p.DelayExp > 15 {
		return nil, errors.New("lorawan: max value of DelayExp is 15")
	}

	return []byte{p.DelayExp | (p.LimitExp << 4)}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *ADRParam) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	p.DelayExp = data[0] & ((1 << 3) | (1 << 2) | (1 << 1) | 1)
	p.LimitExp = (data[0] >> 4)

	return nil
}

// ADRParamSetupReqPayload represents the ADRParamReq payload.
type ADRParamSetupReqPayload struct {
	ADRParam ADRParam `josn:"adrParam"`
}

// MarshalBinary encodes the object into bytes.
func (p ADRParamSetupReqPayload) MarshalBinary() ([]byte, error) {
	return p.ADRParam.MarshalBinary()
}

// UnmarshalBinary decodes the object from bytes.
func (p *ADRParamSetupReqPayload) UnmarshalBinary(data []byte) error {
	return p.ADRParam.UnmarshalBinary(data)
}

// ForceRejoinReqPayload represents the ForceRejoinReq payload.
type ForceRejoinReqPayload struct {
	Period     uint8 `json:"period"`
	MaxRetries uint8 `json:"maxRetries"`
	RejoinType uint8 `json:"rejoinType"`
	DR         uint8 `json:"dr"`
}

// MarshalBinary encodes the object into bytes.
func (p ForceRejoinReqPayload) MarshalBinary() ([]byte, error) {
	if p.Period > 7 {
		return nil, errors.New("lorawan: max value of Period is 7")
	}
	if p.MaxRetries > 7 {
		return nil, errors.New("lorawan: max value of MaxRetries is 7")
	}
	if p.RejoinType != 0 && p.RejoinType != 2 {
		return nil, errors.New("lorawan: RejoinType must be 0 or 2")
	}
	if p.DR > 15 {
		return nil, errors.New("lorawan: max value of DR is 15")
	}

	return []byte{p.DR | (p.RejoinType << 4), p.MaxRetries | (p.Period << 3)}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *ForceRejoinReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 2 {
		return errors.New("lorawan: 2 bytes of data are expected")
	}

	p.DR = data[0] & ((1 << 3) | (1 << 2) | (1 << 1) | 1)
	p.RejoinType = (data[0] & ((1 << 6) | (1 << 5) | (1 << 4))) >> 4
	p.MaxRetries = data[1] & ((1 << 2) | (1 << 1) | 1)
	p.Period = (data[1] & ((1 << 5) | (1 << 4) | (1 << 3))) >> 3

	return nil
}

// RejoinParamSetupReqPayload represents the RejoinParamSetupReq payload.
type RejoinParamSetupReqPayload struct {
	MaxTimeN  uint8 `json:"maxTimeN"`
	MaxCountN uint8 `json:"maxCountN"`
}

// MarshalBinary encodes the object into bytes.
func (p RejoinParamSetupReqPayload) MarshalBinary() ([]byte, error) {
	if p.MaxTimeN > 15 {
		return nil, errors.New("lorawan: max value of MaxTimeN is 15")
	}
	if p.MaxCountN > 15 {
		return nil, errors.New("lorawan: max value of MaxCountN is 15")
	}

	return []byte{p.MaxCountN | (p.MaxTimeN << 4)}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *RejoinParamSetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is exepcted")
	}

	p.MaxCountN = data[0] & ((1 << 3) | (1 << 2) | (1 << 1) | 1)
	p.MaxTimeN = (data[0] & ((1 << 7) | (1 << 6) | (1 << 5) | (1 << 4))) >> 4

	return nil
}

// RejoinParamSetupAnsPayload represents the RejoinParamSetupAns payload.
type RejoinParamSetupAnsPayload struct {
	TimeOK bool `json:"timeOK"`
}

// MarshalBinary encodes the object into bytes.
func (p RejoinParamSetupAnsPayload) MarshalBinary() ([]byte, error) {
	var out byte
	if p.TimeOK {
		out = 1
	}

	return []byte{out}, nil
}

// UnmarshalBinary decodes the object from bytes.
func (p *RejoinParamSetupAnsPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}

	p.TimeOK = data[0]&1 != 0
	return nil
}

// decodeDataPayloadToMACCommands decodes a DataPayload into a slice of
// MACCommands.
func decodeDataPayloadToMACCommands(uplink bool, payloads []Payload) ([]Payload, error) {
	if len(payloads) != 1 {
		return nil, errors.New("lorawan: exactly one Payload expected")
	}

	dataPL, ok := payloads[0].(*DataPayload)
	if !ok {
		return nil, fmt.Errorf("lorawan: expected *DataPayload, got %T", payloads[0])
	}

	var plLen int
	var out []Payload

	for i := 0; i < len(dataPL.Bytes); i++ {
		if _, s, err := GetMACPayloadAndSize(uplink, CID(dataPL.Bytes[i])); err != nil {
			plLen = 0
		} else {
			plLen = s
		}

		if len(dataPL.Bytes[i:]) < plLen+1 {
			return nil, errors.New("lorawan: not enough remaining bytes")
		}

		mc := &MACCommand{}
		if err := mc.UnmarshalBinary(uplink, dataPL.Bytes[i:i+1+plLen]); err != nil {
			log.Printf("warning: unmarshal mac-command error (skipping remaining mac-command bytes): %s", err)
		}

		out = append(out, mc)
		i = i + plLen
	}

	return out, nil
}
