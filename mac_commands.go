package lorawan

import "errors"

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

// ChMask represents the channel mask.
type ChMask [2]byte

// NewChMask returns a new ChMask for the given channel numbers (max. is 16).
func NewChMask(chans ...uint8) (ChMask, error) {
	var mask ChMask
	for _, c := range chans {
		if c > 16 {
			return mask, errors.New("lorawan: the max. channel number is 16")
		}
		c = c - 1 // make it zero based indexed
		i := c / 8
		b := c % 8
		mask[i] = mask[i] ^ 1<<b
	}
	return mask, nil
}

// Channels returns the channels active in the channel mask.
func (m ChMask) Channels() []uint8 {
	var chans []uint8
	for c := uint8(0); c < 16; c++ {
		i := c / 8
		b := c % 8
		if m[i]&(1<<b) > 0 {
			chans = append(chans, c+1)
		}
	}
	return chans
}

// Redundacy represents the redundacy field.
type Redundacy byte

// NewRedundacy returns a new Redundacy. Max allowed value for chMaskCntl is 7,
// max allowed value for nbRep is 15.
func NewRedundacy(chMaskCntl, nbRep uint8) (Redundacy, error) {
	var r Redundacy
	if chMaskCntl > 7 {
		return r, errors.New("lorawan: max value of chMaskCntl is 7")
	}
	if nbRep > 15 {
		return r, errors.New("lorawan: max value of nbRep is 15")
	}

	return Redundacy((chMaskCntl << 4) ^ nbRep), nil
}

// ChMaskCntl (channel mask control) controls the interpretation of the ChMask
// bit field.
func (r Redundacy) ChMaskCntl() uint8 {
	var mask uint8 = (1 << 6) ^ (1 << 5) ^ (1 << 4)
	return (uint8(r) & mask) >> 4
}

// NbRep returns the number of repetition for each uplink message.
func (r Redundacy) NbRep() uint8 {
	var mask uint8 = (1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0)
	return uint8(r) & mask
}

// DataRateTXPower represents the requested data rate and TX output power.
type DataRateTXPower byte

// NewDataRateTXPower returns a new DataRateTXPower. Max allowed values for
// dataRate and txPower are 15.
func NewDataRateTXPower(dataRate, txPower uint8) (DataRateTXPower, error) {
	var dr DataRateTXPower
	if dataRate > 15 {
		return dr, errors.New("lorawan: max value for dataRate is 15")
	}
	if txPower > 15 {
		return dr, errors.New("lorawan: max value for txPower is 15")
	}
	return DataRateTXPower((dataRate << 4) ^ txPower), nil
}

// DataRate returns the requested data rate.
func (dr DataRateTXPower) DataRate() uint8 {
	return uint8(dr) >> 4
}

// TXPower returns the requested TX output power.
func (dr DataRateTXPower) TXPower() uint8 {
	var mask uint8 = (1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0)
	return uint8(dr) & mask
}

// LinkADRReqPayload represents the LinkADRReq payload.
type LinkADRReqPayload struct {
	DataRateTXPower DataRateTXPower
	ChMask          ChMask
	Redundacy       Redundacy
}
