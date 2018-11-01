package lorawan

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type macCommandPayloadTest struct {
	Payload MACCommandPayload
	Bytes   []byte
	Error   error
}

type MACCommandPayloadTestSuite struct {
	suite.Suite
}

func (ts MACCommandPayloadTestSuite) run(newPLFunc func() MACCommandPayload, tests []macCommandPayloadTest) {
	assert := require.New(ts.T())

	for _, tst := range tests {
		if tst.Payload != nil {
			b, err := tst.Payload.MarshalBinary()
			assert.NoError(err)
			assert.Equal(tst.Bytes, b)
		}

		pl := newPLFunc()
		err := pl.UnmarshalBinary(tst.Bytes)
		if tst.Error == nil {
			assert.NoError(err)
		} else {
			assert.Equal(tst.Error, err)
		}

		if err == nil {
			assert.Equal(tst.Payload, pl)
		}
	}
}

func (ts MACCommandPayloadTestSuite) TestDeviceModeIndClass() {
	tests := []macCommandPayloadTest{
		{
			Bytes: []byte{},
			Error: errors.New("lorawan: 1 byte of data is expected"),
		},
		{
			Bytes: []byte{0x00, 0x01},
			Error: errors.New("lorawan: 1 byte of data is expected"),
		},
		{
			Payload: &DeviceModeIndPayload{
				Class: DeviceModeIndClassA,
			},
			Bytes: []byte{0x00},
		},
		{
			Payload: &DeviceModeIndPayload{
				Class: DeviceModeIndClassC,
			},
			Bytes: []byte{0x02},
		},
	}

	ts.run(func() MACCommandPayload { return &DeviceModeIndPayload{} }, tests)
}

func TestMACCommandPayloads(t *testing.T) {
	suite.Run(t, new(MACCommandPayloadTestSuite))
}
