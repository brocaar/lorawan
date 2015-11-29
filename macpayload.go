package lorawan

import "errors"

// MACPayload represents the MAC payload.
type MACPayload struct {
	FHDR       FHDR
	FPort      *uint8
	FRMPayload []Payload
	uplink     bool // used for binary (un)marshaling
}

func (p MACPayload) MarshalBinary() ([]byte, error) {
	var b []byte
	var err error
	out := make([]byte, 0)

	p.FHDR.uplink = p.uplink
	b, err = p.FHDR.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	out = append(out, b...)

	if len(p.FRMPayload) == 0 {
		if p.FPort != nil {
			return []byte{}, errors.New("lorawan: FPort should not be set when FRMPayload is empty")
		}
		return out, nil
	}

	out = append(out, *p.FPort)
	for _, fp := range p.FRMPayload {
		if mac, ok := fp.(*MACCommand); ok {
			if *p.FPort != 0 {
				return []byte{}, errors.New("lorawan: a MAC command is only allowed when FPort=0")
			}
			mac.uplink = p.uplink
			b, err = mac.MarshalBinary()
		} else {
			b, err = fp.MarshalBinary()
		}
		if err != nil {
			return []byte{}, err
		}
		out = append(out, b...)
	}

	return out, nil
}

func (p *MACPayload) UnmarshalBinary(data []byte) error {
	// check that there are enough bytes to decode a minimal FHDR
	if len(data) < 7 {
		return errors.New("lorawan: at least 7 bytes needed to decode FHDR")
	}

	// unmarshal FCtrl so we know the FOptsLen
	if err := p.FHDR.FCtrl.UnmarshalBinary(data[4:5]); err != nil {
		return err
	}

	// check that there are at least as many bytes as FOptsLen claims
	if len(data) < 7+int(p.FHDR.FCtrl.fOptsLen) {
		return errors.New("lorawan: not enough bytes to decode FHDR")
	}

	// decode the full FHDR (including optional FOpts)
	if err := p.FHDR.UnmarshalBinary(data[0 : 7+p.FHDR.FCtrl.fOptsLen]); err != nil {
		return err
	}

	// check that there are at least 2 more bytes (FPort and FRMPayload)
	if len(data) < 7+int(p.FHDR.FCtrl.fOptsLen)+2 {
		if len(data) == 7+int(p.FHDR.FCtrl.fOptsLen)+1 {
			return errors.New("lorawan: data contains FPort but no FRMPayload")
		}
		return nil
	}

	fPort := uint8(data[7+p.FHDR.FCtrl.fOptsLen])
	p.FPort = &fPort
	payload := data[7+p.FHDR.FCtrl.fOptsLen+1:]

	if *p.FPort == 0 {
		// payload contains MAC commands
		var pLen int
		for i := 0; i < len(payload); i++ {
			if _, s, err := getMACPayloadAndSize(p.uplink, cid(payload[i])); err != nil {
				pLen = 0
			} else {
				pLen = s
			}

			// check if the remaining bytes are >= CID byte + payload size
			if len(payload[i:]) < pLen+1 {
				return errors.New("lorawan: not enough remaining bytes")
			}

			mc := &MACCommand{uplink: p.uplink}
			if err := mc.UnmarshalBinary(payload[i : i+1+pLen]); err != nil {
				return err
			}
			p.FRMPayload = append(p.FRMPayload, mc)

			// go to the next command (skip the payload bytes of the current command)
			i = i + pLen
		}

	} else {
		// payload contains user defined data
		p.FRMPayload = []Payload{&DataPayload{}}
		if err := p.FRMPayload[0].UnmarshalBinary(payload); err != nil {
			return err
		}
	}

	return nil
}
