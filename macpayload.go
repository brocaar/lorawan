package lorawan

// MACPayload represents the MAC payload.
type MACPayload struct {
	FHDR       FHDR
	FPort      *uint8
	FRMPayload []byte
}
