package firmwaremanagement

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFragmentation(t *testing.T) {
	nextFirmwareVersion := uint32(262657)

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
			Name:   "PackageVersionAns",
			Uplink: true,
			Command: Command{
				CID: PackageVersionAns,
				Payload: &PackageVersionAnsPayload{
					PackageIdentifier: 1,
					PackageVersion:    1,
				},
			},
			Bytes: []byte{0x00, 0x01, 0x01},
		},
		{
			Name:                   "PackageVersionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x00, 0x01},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 2 bytes are expected"),
		},
		{
			Name: "DevVersionReq",
			Command: Command{
				CID:     DevVersionReq,
				Payload: &DevVersionReqPayload{},
			},
			Bytes: []byte{0x01},
		},
		{
			Name:                   "DevVersionReq invalid bytes",
			Bytes:                  []byte{0x01, 0x2},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 0 bytes are expected"),
		},
		{
			Name:   "DevVersionAns",
			Uplink: true,
			Command: Command{
				CID: DevVersionAns,
				Payload: &DevVersionAnsPayload{
					FWversion: 513,
					HWversion: 1264,
				},
			},
			Bytes: []byte{0x01, 0x01, 0x02, 0x00, 0x00, 0xF0, 0x04, 0x00, 0x00},
		},
		{
			Name:                   "DevVersionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x01, 0x02, 0x01, 0x04},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 8 bytes are expected"),
		},
		{
			Name: "DevRebootTimeReq",
			Command: Command{
				CID: DevRebootTimeReq,
				Payload: &DevRebootTimeReqPayload{
					RebootTime: 134480385,
				},
			},
			Bytes: []byte{0x02, 0x01, 0x02, 0x04, 0x08},
		},
		{
			Name:                   "DevRebootTimeReq invalid bytes",
			Bytes:                  []byte{0x02, 0x01, 0x02, 0x04},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 4 bytes are expected"),
		},
		{
			Name:   "DevRebootTimeAns",
			Uplink: true,
			Command: Command{
				CID: DevRebootTimeAns,
				Payload: &DevRebootTimeAnsPayload{
					RebootTime: 134480385,
				},
			},
			Bytes: []byte{0x02, 0x01, 0x02, 0x04, 0x08},
		},
		{
			Name:                   "DevRebootTimeAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x02, 0x01, 0x02, 0x04},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 4 bytes are expected"),
		},
		{
			Name: "DevRebootCountdownReq",
			Command: Command{
				CID: DevRebootCountdownReq,
				Payload: &DevRebootCountdownReqPayload{
					Countdown: 262657,
				},
			},
			Bytes: []byte{0x03, 0x01, 0x02, 0x04},
		},
		{
			Name:                   "DevRebootCountdownReq invalid bytes",
			Bytes:                  []byte{0x03, 0x01, 0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 3 bytes are expected"),
		},
		{
			Name:   "DevRebootCountdownAns",
			Uplink: true,
			Command: Command{
				CID: DevRebootCountdownAns,
				Payload: &DevRebootCountdownAnsPayload{
					Countdown: 262657,
				},
			},
			Bytes: []byte{0x03, 0x01, 0x02, 0x04},
		},
		{
			Name:                   "DevRebootCountdownAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x03, 0x01, 0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 3 bytes are expected"),
		},
		{
			Name: "DevUpgradeImageReq",
			Command: Command{
				CID:     DevUpgradeImageReq,
				Payload: &DevUpgradeImageReqPayload{},
			},
			Bytes: []byte{0x04},
		},
		{
			Name:   "DevUpgradeImageAns Valid Firmware",
			Uplink: true,
			Command: Command{
				CID: DevUpgradeImageAns,
				Payload: &DevUpgradeImageAnsPayload{
					Status: DevUpgradeImageAnsPayloadStatus{
						UpImageStatus: FirmwareValid,
					},
					nextFirmwareVersion: &nextFirmwareVersion,
				},
			},
			Bytes: []byte{0x04, 0x03, 0x01, 0x02, 0x04, 0x00},
		},
		{
			Name:   "DevUpgradeImageAns Invalid Firmware",
			Uplink: true,
			Command: Command{
				CID: DevUpgradeImageAns,
				Payload: &DevUpgradeImageAnsPayload{
					Status: DevUpgradeImageAnsPayloadStatus{
						UpImageStatus: FirmwareCorruptOrInvalidSignature,
					},
				},
			},
			Bytes: []byte{0x04, 0x01}},

		{
			Name:   "DevUpgradeImageAns InvalidFirmware with nextFirmwareVersion",
			Uplink: true,
			Command: Command{
				CID: DevUpgradeImageAns,
				Payload: &DevUpgradeImageAnsPayload{
					Status: DevUpgradeImageAnsPayloadStatus{
						UpImageStatus: FirmwareCorruptOrInvalidSignature,
					},
					nextFirmwareVersion: &nextFirmwareVersion,
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/firmwaremanagement: nextFirmwareVersion must be nil when UpImageStatus != 3 due no valid firmware present"),
		},
		{
			Name:                   "DevVersionAns invalid bytes - Valid Firware in UplinkStatus",
			Uplink:                 true,
			Bytes:                  []byte{0x04, 0x03, 0x01, 0x02, 0x03},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 5 bytes are expected"),
		},
		{
			Name:                   "DevVersionAns invalid bytes - no UplinkStatus",
			Uplink:                 true,
			Bytes:                  []byte{0x04},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: at least 1 byte is expected"),
		},
		{
			Name: "DevDeleteImageReq",
			Command: Command{
				CID: DevDeleteImageReq,
				Payload: &DevDeleteImageReqPayload{
					FirmwareToDeleteVersion: 197121,
				},
			},
			Bytes: []byte{0x05, 0x01, 0x02, 0x03, 0x00},
		},
		{
			Name:                   "DevVersionReq invalid bytes",
			Bytes:                  []byte{0x05, 0x01, 0x02, 0x0},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 4 bytes are expected"),
		},
		{
			Name:   "DevDeleteImageAns",
			Uplink: true,
			Command: Command{
				CID: DevDeleteImageAns,
				Payload: &DevDeleteImageAnsPayload{
					Status: DevDeleteImageAnsStatus{
						ErrorInvalidVersion: 1,
						ErrorNoValidImage:   1,
					},
				},
			},
			Bytes: []byte{0x05, 0x3},
		},
		{
			Name:                   "DevVersionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x05},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/firmwaremanagement: 1 bytes are expected"),
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
