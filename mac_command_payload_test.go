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
			assert.Equal(tst.Error, err)
			assert.Equal(tst.Bytes, b)

			// if there is a Payload and error, skip to the next test
			if tst.Error != nil {
				continue
			}
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
				Class: DeviceModeClassA,
			},
			Bytes: []byte{0x00},
		},
		{
			Payload: &DeviceModeIndPayload{
				Class: DeviceModeClassC,
			},
			Bytes: []byte{0x02},
		},
	}

	ts.run(func() MACCommandPayload { return &DeviceModeIndPayload{} }, tests)
}

func (ts MACCommandPayloadTestSuite) TestDeviceModeConfClass() {
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
			Payload: &DeviceModeConfPayload{
				Class: DeviceModeClassA,
			},
			Bytes: []byte{0x00},
		},
		{
			Payload: &DeviceModeConfPayload{
				Class: DeviceModeClassC,
			},
			Bytes: []byte{0x02},
		},
	}

	ts.run(func() MACCommandPayload { return &DeviceModeConfPayload{} }, tests)
}

func (ts MACCommandPayloadTestSuite) TestTXParamSetupReqPayload() {
	tests := []macCommandPayloadTest{
		{
			Bytes: []byte{},
			Error: errors.New("lorawan: 1 byte of data is expected"),
		},
		{
			Payload: &TXParamSetupReqPayload{
				UplinkDwellTime:   DwellTime400ms,
				DownlinkDwelltime: DwellTimeNoLimit,
				MaxEIRP:           7,
			},
			Bytes: []byte{0x17},
		},
		{
			Payload: &TXParamSetupReqPayload{
				UplinkDwellTime:   DwellTimeNoLimit,
				DownlinkDwelltime: DwellTime400ms,
				MaxEIRP:           7,
			},
			Bytes: []byte{0x27},
		},
		{
			Payload: &TXParamSetupReqPayload{
				UplinkDwellTime:   DwellTimeNoLimit,
				DownlinkDwelltime: DwellTime400ms,
				MaxEIRP:           16,
			},
			Error: errors.New("lorawan: max value of MaxEIRP is 15"),
		},
	}

	ts.run(func() MACCommandPayload { return &TXParamSetupReqPayload{} }, tests)
}

func TestMACCommandPayloads(t *testing.T) {
	suite.Run(t, new(MACCommandPayloadTestSuite))
}
