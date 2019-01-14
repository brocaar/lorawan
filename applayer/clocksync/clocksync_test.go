package clocksync

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClockSync(t *testing.T) {
	tests := []struct {
		Name                   string
		Command                Command
		Bytes                  []byte
		Uplink                 bool
		ExpectedMarshalError   error
		ExpectedUnmarshalError error
	}{
		{
			Name: "PackageVersionReq",
			Command: Command{
				CID: PackageVersionReq,
			},
			Bytes: []byte{0x00},
		},
		{
			Name: "PackageVersionAns",
			Command: Command{
				CID: PackageVersionAns,
				Payload: &PackageVersionAnsPayload{
					PackageIdentifier: 1,
					PackageVersion:    1,
				},
			},
			Uplink: true,
			Bytes:  []byte{0x00, 0x01, 0x01},
		},
		{
			Name:                   "PackageVersionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x00, 0x01},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 2 bytes are expected"),
		},
		{
			Name:   "AppTimeReq",
			Uplink: true,
			Command: Command{
				CID: AppTimeReq,
				Payload: &AppTimeReqPayload{
					DeviceTime: 134480385,
					Param: AppTimeReqPayloadParam{
						TokenReq:    5,
						AnsRequired: true,
					},
				},
			},
			Bytes: []byte{0x01, 0x01, 0x02, 0x04, 0x08, 0x15},
		},
		{
			Name:                   "AppTimeReq invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x01, 0x01, 0x02, 0x04, 0x08},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 5 bytes are expected"),
		},
		{
			Name: "AppTimeAns",
			Command: Command{
				CID: AppTimeAns,
				Payload: &AppTimeAnsPayload{
					TimeCorrection: -134480385,
					Param: AppTimeAnsPayloadParam{
						TokenAns: 5,
					},
				},
			},
			Bytes: []byte{0x01, 0xff, 0xfd, 0xfb, 0xf7, 0x05},
		},
		{
			Name:                   "AppTimeAns invalid bytes",
			Bytes:                  []byte{0x01, 0x01, 0x02, 0x04, 0x08},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 5 bytes are expected"),
		},
		{
			Name: "DeviceAppTimePeriodicityReq",
			Command: Command{
				CID: DeviceAppTimePeriodicityReq,
				Payload: &DeviceAppTimePeriodicityReqPayload{
					Periodicity: DeviceAppTimePeriodicityReqPayloadPeriodicity{
						5,
					},
				},
			},
			Bytes: []byte{0x02, 0x05},
		},
		{
			Name:                   "DeviceAppTimePeriodicityReq invalid bytes",
			Bytes:                  []byte{0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 1 bytes are expected"),
		},
		{
			Name:   "DeviceAppTimePeriodicityAns",
			Uplink: true,
			Command: Command{
				CID: DeviceAppTimePeriodicityAns,
				Payload: &DeviceAppTimePeriodicityAnsPayload{
					Status: DeviceAppTimePeriodicityAnsPayloadStatus{
						NotSupported: true,
					},
					Time: 134480385,
				},
			},
			Bytes: []byte{0x02, 0x01, 0x01, 0x02, 0x04, 0x08},
		},
		{
			Name:                   "DeviceAppTimePeriodicityAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x02, 0x01, 0x01, 0x02, 0x04},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 5 bytes are expected"),
		},
		{
			Name: "ForceDeviceResyncReq",
			Command: Command{
				CID: ForceDeviceResyncReq,
				Payload: &ForceDeviceResyncReqPayload{
					ForceConf: ForceDeviceResyncReqPayloadForceConf{
						NbTransmissions: 5,
					},
				},
			},
			Bytes: []byte{0x03, 0x05},
		},
		{
			Name:                   "ForceDeviceResyncReq invalid bytes",
			Bytes:                  []byte{0x03},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/clocksync: 1 bytes are expected"),
		},
	}

	for _, tst := range tests {
		t.Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			if tst.ExpectedMarshalError != nil {
				_, err := tst.Command.MarshalBinary()
				assert.Equal(tst.ExpectedMarshalError, err)
			} else if tst.ExpectedUnmarshalError != nil {
				var cmd Command
				err := cmd.UnmarshalBinary(tst.Uplink, tst.Bytes)
				assert.Equal(tst.ExpectedUnmarshalError, err)
			} else {
				cmds := Commands{tst.Command}
				b, err := cmds.MarshalBinary()
				assert.NoError(err)
				assert.Equal(tst.Bytes, b)

				cmds = Commands{}
				assert.NoError(cmds.UnmarshalBinary(tst.Uplink, tst.Bytes))
				assert.Len(cmds, 1)
				assert.Equal(tst.Command, cmds[0])
			}
		})
	}
}
