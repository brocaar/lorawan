package lorawan

import "errors"

// DevAddr represents the device address.
type DevAddr [4]byte

// FCtrl represents the frame control field.
type FCtrl byte

// NewFCtrl returns a new FCtrl. Note that for fOptsLen only the first
// four bits are used (and thus the max. allowed number is 15)
func NewFCtrl(adr, adrAckReq, ack, fPending bool, fOptsLen uint8) (FCtrl, error) {
	var fc FCtrl
	if fOptsLen > 15 {
		return fc, errors.New("lorawan: the max. fOptsLen is 15")
	}

	if adr {
		fc = fc ^ (1 << 7)
	}
	if adrAckReq {
		fc = fc ^ (1 << 6)
	}
	if ack {
		fc = fc ^ (1 << 5)
	}
	if fPending {
		fc = fc ^ (1 << 4)
	}

	return fc ^ FCtrl(fOptsLen), nil
}

// ADR returns if the adaptive data rate control bit is set.
func (c FCtrl) ADR() bool {
	return c&(1<<7) > 0
}

// ADRACKReq returns if the acknowledgment request bit is set.
func (c FCtrl) ADRACKReq() bool {
	return c&(1<<6) > 0
}

// ACK returns if the acknowledgment bit is set.
func (c FCtrl) ACK() bool {
	return c&(1<<5) > 0
}

// FPending returns if the gataway has more data pending to be sent.
// This is only used in downlink communication.
func (c FCtrl) FPending() bool {
	return c&(1<<4) > 0
}

// FOptsLen returns how many FOpts bytes the FHDR has.
func (c FCtrl) FOptsLen() uint8 {
	var mask uint8 = (1 << 3) ^ (1 << 2) ^ (1 << 1) ^ (1 << 0)
	return uint8(c) & mask
}

// FHDR represents the frame header.
type FHDR struct {
	DevAddr DevAddr
	FCtrl   FCtrl
	Fcnt    uint16
	FOpts   []byte // max. number of allowed bytes is 15
}
