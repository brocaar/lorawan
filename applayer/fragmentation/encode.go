package fragmentation

import (
	"errors"
)

// Encode encodes the given slice of bytes to fragments including forward error correction.
// This is based on the proposed FEC code from the Fragmented Data Block Transport over
// LoRaWAN recommendation.
func Encode(data []byte, fragmentSize, redundancy int) ([][]byte, error) {
	if len(data)%fragmentSize != 0 {
		return nil, errors.New("length of data must be a multiple of the given fragment-size")
	}

	// fragment the data into rows
	var dataRows [][]byte
	for i := 0; i < len(data)/fragmentSize; i++ {
		offset := i * fragmentSize
		dataRows = append(dataRows, data[offset:offset+fragmentSize])
	}
	w := len(dataRows)

	for y := 0; y < redundancy; y++ {
		s := make([]byte, fragmentSize)
		a := matrixLine(y+1, w)

		for x := 0; x < w; x++ {
			if a[x] == 1 {
				for m := 0; m < fragmentSize; m++ {
					s[m] ^= dataRows[x][m]
				}
			}
		}

		dataRows = append(dataRows, s)
	}

	return dataRows, nil
}

func prbs23(x int) int {
	b0 := x & 1
	b1 := (x & 32) / 32
	return (x / 2) + (b0^b1)*(1<<22)
}

func isPower2(num int) bool {
	return num != 0 && (num&(num-1)) == 0
}

func matrixLine(n, m int) []int {
	line := make([]int, m)

	mm := 0
	if isPower2(m) {
		mm = 1
	}

	x := 1 + (1001 * n)

	for nbCoeff := 0; nbCoeff < m/2; nbCoeff++ {
		r := 1 << 16
		for r >= m {
			x = prbs23(x)
			r = x % (m + mm)
		}
		line[r] = 1
	}

	return line
}
