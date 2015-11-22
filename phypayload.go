package lorawan

// PHYPayload represents the physical payload.
type PHYPayload struct {
	MHDR       MHDR
	MACPayload MACPayload
	MIC        [4]byte
}

// CalculateMIC calculates and returns the integrity code for the payload.
func (p PHYPayload) CalculateMIC() [4]byte {
	panic("not implemented")
}
