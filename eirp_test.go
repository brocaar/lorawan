package lorawan

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTXParamSetupEIRPIndex(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		EIRP          float32
		ExpectedIndex uint8
	}{
		{8, 0},
		{9, 0},
		{10, 1},
		{36, 15},
		{37, 15},
		{12.15, 2},
	}

	for _, tst := range tests {
		assert.Equal(tst.ExpectedIndex, GetTXParamSetupEIRPIndex(tst.EIRP))
	}
}

func TestGetTXParamSetupEIRP(t *testing.T) {
	assert := require.New(t)

	tests := []struct {
		Index uint8
		EIRP  float32
		Error error
	}{
		{0, 8, nil},
		{15, 36, nil},
		{16, 0, errors.New("lorawan: invalid eirp index")},
	}

	for _, tst := range tests {
		e, err := GetTXParamSetupEIRP(tst.Index)
		assert.Equal(tst.Error, err)

		if err == nil {
			assert.Equal(tst.EIRP, e)
		}
	}
}
