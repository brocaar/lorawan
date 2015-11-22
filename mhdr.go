package lorawan

// MType represents the message type.
type MType byte

// Major defines the major version of data message.
type Major byte

// Supported message types (MType)
const (
	JoinRequest         MType = 0
	JoinAccept          MType = (1 << 5)
	UnconfirmedDataUp   MType = (1 << 6)
	UnconfirmedDataDown MType = (1 << 6) ^ (1 << 5)
	ConfirmedDataUp     MType = (1 << 7)
	ConfirmedDataDown   MType = (1 << 7) ^ (1 << 5)
	MTypeRFU            MType = (1 << 7) ^ (1 << 6)
	Proprietary         MType = (1 << 7) ^ (1 << 6) ^ (1 << 5)
)

// Supported major versions
const (
	LoRaWANR1 Major = 0
	MajorRFU1 Major = (1 << 0)
	MajorRFU2 Major = (1 << 1)
	MajorRFU3 Major = (1 << 1) ^ (1 << 0)
)

// MHDR represents the MAC header field.
type MHDR byte

// NewMHDR returns a new MAC header for the given type and major version.
func NewMHDR(mtype MType, major Major) MHDR {
	return MHDR(byte(mtype) ^ byte(major))
}

// MType returns the message type.
func (h MHDR) MType() MType {
	var mask MType = (1 << 7) ^ (1 << 6) ^ (1 << 5)
	return MType(h) & mask
}

// Major returns the major version.
func (h MHDR) Major() Major {
	var mask Major = (1 << 1) ^ (1 << 0)
	return Major(h) & mask
}
