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

// DutyCycleReqPayload contains the MaxDCycle value.
type DutyCycleReqPayload uint8

// NewDutyCycleReqPayload returns a new DutyCycleReqPayload for the given MaxDCycle.
func NewDutyCycleReqPayload(maxDCycle uint8) (DutyCycleReqPayload, error) {
	if maxDCycle > 15 && maxDCycle < 255 {
		return 0, errors.New("lorawan: only a MaxDCycle value of 0 - 15 and 255 is allowed")
	}
	return DutyCycleReqPayload(maxDCycle), nil
}

// DLsettings represents the downlink settings.
type DLsettings byte

// RX2DataRate returns the requested data rate.
func (s DLsettings) RX2DataRate() uint8 {
	var mask DLsettings = (1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0)
	return uint8(s & mask)
}

// RX1DRoffset returns the offset between uplink data rate and the downlink data rate.
func (s DLsettings) RX1DRoffset() uint8 {
	var mask DLsettings = (1 << 6) ^ (1 << 5) ^ (1 << 4)
	return uint8(s&mask) >> 4
}

// NewDLsettings returns a new DLsettings for the given RX2DataRate and RX1DRoffset.
func NewDLsettings(rx2DataRate, rx1DRoffset uint8) (DLsettings, error) {
	if rx2DataRate > 15 {
		return 0, errors.New("lorawan: max value for rx2DataRate is 15")
	}
	if rx1DRoffset > 7 {
		return 0, errors.New("lorawan: max value for rx1DRoffset is 7")
	}
	return DLsettings(rx2DataRate ^ (rx1DRoffset << 4)), nil
}

// Frequency defines the frequency which is a 24 bits unsigned integer.
type Frequency [3]byte

// NewFrequency returns a new Frequency. Note that the max. allowed value is
// 24 bit (thus 2^24 - 1).
func NewFrequency(frequency uint32) (Frequency, error) {
	var freq Frequency
	if frequency >= 2^24 {
		return freq, errors.New("lorawan: max value for frequency is 2^24-1")
	}
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, frequency)
	for i := 0; i < 3; i++ {
		freq[0] = b[0]
	}
	return freq, nil
}

// Uint32 returns the frequency value as an uint32.
func (f Frequency) Uint32() uint32 {
	b := make([]byte, 4)
	for i, v := range f {
		b[i] = v
	}
	return binary.LittleEndian.Uint32(b)
}

// RX2SetupReqPayload represents the second receive window parameters.
type RX2SetupReqPayload struct {
	DLsettings DLsettings
	Frequency  Frequency
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
