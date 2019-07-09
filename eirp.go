package lorawan

import "errors"

var eirpTable = [...]float32{
	8,  // 0
	10, // 1
	12, // 2
	13, // 3
	14, // 4
	16, // 5
	18, // 6
	20, // 7
	21, // 8
	24, // 9
	26, // 10
	27, // 11
	29, // 12
	30, // 13
	33, // 14
	36, // 15
}

// GetTXParamSetupEIRPIndex returns the coded value for the given EIRP (dBm).
// Note that it returns the coded value that is closest to the given EIRP,
// without exceeding it.
func GetTXParamSetupEIRPIndex(eirp float32) uint8 {
	var out uint8
	for i, e := range eirpTable {
		if e > eirp {
			return out
		}
		out = uint8(i)
	}
	return out
}

// GetTXParamsetupEIRP returns the EIRP (dBm) for the coded value.
func GetTXParamSetupEIRP(index uint8) (float32, error) {
	if int(index) > len(eirpTable)-1 {
		return 0, errors.New("lorawan: invalid eirp index")
	}
	return eirpTable[index], nil
}
