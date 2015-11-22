package lorawan

// DevAddr represents the device address.
type DevAddr [4]byte

// FCtrl represents the frame control field.
type FCtrl byte

// MACPayload represents the MAC payload.
type MACPayload struct {
	DevAddr DevAddr
	FCtrl   FCtrl
	FCnt    uint16
	FOpts   []byte // maximun number of allowed bytes is 15
}
