package lorawan

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// NetID represents the NetID.
type NetID [3]byte

// String implements fmt.Stringer.
func (n NetID) String() string {
	return hex.EncodeToString(n[:])
}

// MarshalJSON implements json.Marshaler.
func (n NetID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + n.String() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NetID) UnmarshalJSON(data []byte) error {
	hexStr := strings.Trim(string(data), `"`)
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	if len(b) != len(n) {
		return fmt.Errorf("lorawan: exactly %d bytes are expected", len(n))
	}
	copy(n[:], b)
	return nil
}

// NwkID returns the seven LSB of the NetID.
func (n NetID) NwkID() byte {
	return n[2] & 127 // 7 lsb
}
