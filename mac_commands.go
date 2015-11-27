package lorawan

import (
	"encoding/binary"
	"errors"
)

// CID defines the MAC command identifier.
type CID byte

// MAC commands as specified by the LoRaWAN R1.0 specs. Note that each *Req / *Ans
// has the same value. Based on the fact if a message is uplink or downlink
// you should use on or the other.
const (
	LinkCheckReq     CID = 0x02
	LinkCheckAns     CID = 0x02
	LinkADRReq       CID = 0x03
	LinkADRAns       CID = 0x03
	DutyCycleReq     CID = 0x04
	DutyCycleAns     CID = 0x04
	RXParamSetupReq  CID = 0x05
	RXParamSetupAns  CID = 0x05
	DevStatusReq     CID = 0x06
	DevStatusAns     CID = 0x06
	NewChannelReq    CID = 0x07
	NewChannelAns    CID = 0x07
	RXTimingSetupReq CID = 0x08
	RXTimingSetupAns CID = 0x08
	// 0x80 to 0xFF reserved for proprietary network command extensions
)

// LinkCheckAnsPayload represents the LinkCheckAns payload.
type LinkCheckAnsPayload struct {
	Margin uint8
	GwCnt  uint8
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
	ChMaskCntl uint8
	NbRep      uint8
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
	DataRate   uint8
	TXPower    uint8
	ChMask     ChMask
	Redundancy Redundancy
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
	if err := p.Redundancy.UnmarshalBinary(data[3:4]); err != nil {
		return err
	}
	return nil
}

// LinkADRAnsPayload represents the LinkADRAns payload.
type LinkADRAnsPayload struct {
	ChannelMaskACK bool
	DataRateACK    bool
	PowerACK       bool
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
	MaxDCCycle uint8
}

// MarshalBinary marshals the object in binary form.
func (p DutyCycleReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 1)
	if p.MaxDCCycle > 15 && p.MaxDCCycle < 255 {
		return b, errors.New("lorawan: only a MaxDCycle value of 0 - 15 and 255 is allowed")
	}
	b = append(b, p.MaxDCCycle)
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (p *DutyCycleReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	p.MaxDCCycle = data[0]
	return nil
}

// DLsettings represents the DLsettings fields (downlink settings).
type DLsettings struct {
	RX2DataRate uint8
	RX1DRoffset uint8
}

// MarshalBinary marshals the object in binary form.
func (s DLsettings) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, 1)
	if s.RX2DataRate > 15 {
		return b, errors.New("lorawan: max value of RX2DataRate is 15")
	}
	if s.RX1DRoffset > 7 {
		return b, errors.New("lorawan: max value of RX1DRoffset is 7")
	}
	b = append(b, s.RX2DataRate^(s.RX1DRoffset<<4))
	return b, nil
}

// UnmarshalBinary decodes the object from binary form.
func (s *DLsettings) UnmarshalBinary(data []byte) error {
	if len(data) != 1 {
		return errors.New("lorawan: 1 byte of data is expected")
	}
	s.RX2DataRate = data[0] & ((1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0))
	s.RX1DRoffset = (data[0] & ((1 << 6) ^ (1 << 5) ^ (1 << 4))) >> 4
	return nil
}

// RX2SetupReqPayload represents the RX2SetupReq payload.
type RX2SetupReqPayload struct {
	Frequency  uint32
	DLsettings DLsettings
}

// MarshalBinary marshals the object in binary form.
func (p RX2SetupReqPayload) MarshalBinary() ([]byte, error) {
	b := make([]byte, 5)
	if p.Frequency >= 16777216 { // 2^24
		return b, errors.New("lorawan: max value of Frequency is 2^24-1")
	}
	if bytes, err := p.DLsettings.MarshalBinary(); err != nil {
		return b, err
	} else {
		b[0] = bytes[0]
	}
	binary.LittleEndian.PutUint32(b[1:5], p.Frequency)
	// we don't return the last octet which is fine since we're only interested
	// in the 24 LSB of Frequency
	return b[0:4], nil
}

func (p *RX2SetupReqPayload) UnmarshalBinary(data []byte) error {
	if len(data) != 4 {
		return errors.New("lorawan: 4 bytes of data are expected")
	}
	if err := p.DLsettings.UnmarshalBinary(data[0:1]); err != nil {
		return err
	}
	// append one block of empty bits at the end of the slice since the
	// binary to uint32 expects 32 bits.
	data = append(data, byte(0))
	p.Frequency = binary.LittleEndian.Uint32(data[1:5])
	return nil
}

// RX2SetupAnsPayload represents payload send by the RXParamSetupAns command.
type RX2SetupAnsPayload byte

// NewRX2SetupAnsPayload returns a new RX2SetupAnsPayload.
func NewRX2SetupAnsPayload(channelACK, rx2DataRateACK, rx1DRoffsetACK bool) RX2SetupAnsPayload {
	var p RX2SetupAnsPayload
	if channelACK {
		p = p ^ (1 << 0)
	}
	if rx2DataRateACK {
		p = p ^ (1 << 1)
	}
	if rx1DRoffsetACK {
		p = p ^ (1 << 2)
	}
	return p
}

// ChannelACK returns if the RX2 slot was successfully set.
func (p RX2SetupAnsPayload) ChannelACK() bool {
	return p&(1<<0) > 0
}

// RX2DataRateACK returns if the RX2 slot data rate was successfully set.
func (p RX2SetupAnsPayload) RX2DataRateACK() bool {
	return p&(1<<1) > 0
}

// RX1DRoffsetACK return if the RX1DRoffset was successfully set.
func (p RX2SetupAnsPayload) RX1DRoffsetACK() bool {
	return p&(1<<2) > 0
}
