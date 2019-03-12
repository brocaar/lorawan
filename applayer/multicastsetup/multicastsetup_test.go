package multicastsetup

import (
	"errors"
	"testing"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/require"
)

func TestMulticastSetup(t *testing.T) {
	timeToStart := uint32(262657)

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
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 2 bytes are expected"),
		},
		{
			Name: "McGroupStatusReq",
			Command: Command{
				CID: McGroupStatusReq,
				Payload: &McGroupStatusReqPayload{
					CmdMask: McGroupStatusReqPayloadCmdMask{
						RegGroupMask: [4]bool{
							false,
							true,
							false,
							true,
						},
					},
				},
			},
			Bytes: []byte{0x01, 0x0a},
		},
		{
			Name:   "McGroupStatusAns",
			Uplink: true,
			Command: Command{
				CID: McGroupStatusAns,
				Payload: &McGroupStatusAnsPayload{
					Status: McGroupStatusAnsPayloadStatus{
						NbTotalGroups: 4,
						AnsGroupMask:  [4]bool{false, false, true, true},
					},
					Items: []McGroupStatusAnsPayloadItem{
						{
							McGroupID: 3,
							McAddr:    lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
						},
						{
							McGroupID: 2,
							McAddr:    lorawan.DevAddr{0x04, 0x03, 0x02, 0x01},
						},
					},
				},
			},
			Bytes: []byte{0x01, 0x4c, 0x03, 0x04, 0x03, 0x02, 0x01, 0x02, 0x01, 0x02, 0x03, 0x04},
		},
		{
			Name:                   "McGroupStatusAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x01, 0x4c, 0x03, 0x04, 0x03, 0x02, 0x01, 0x02, 0x01, 0x02, 0x03},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 11 bytes are expected"),
		},
		{
			Name:   "McGroupStatusAns invalid item count",
			Uplink: true,
			Command: Command{
				CID: McGroupStatusAns,
				Payload: &McGroupStatusAnsPayload{
					Status: McGroupStatusAnsPayloadStatus{
						NbTotalGroups: 4,
						AnsGroupMask:  [4]bool{false, false, false, true},
					},
					Items: []McGroupStatusAnsPayloadItem{
						{
							McGroupID: 3,
							McAddr:    lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
						},
						{
							McGroupID: 2,
							McAddr:    lorawan.DevAddr{0x04, 0x03, 0x02, 0x01},
						},
					},
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/multicastsetup: number of items does not match AnsGroupMatch"),
		},
		{
			Name: "McGroupSetupReq",
			Command: Command{
				CID: McGroupSetupReq,
				Payload: &McGroupSetupReqPayload{
					McGroupIDHeader: McGroupSetupReqPayloadMcGroupIDHeader{
						McGroupID: 2,
					},
					McAddr:         lorawan.DevAddr{0x01, 0x02, 0x03, 0x04},
					McKeyEncrypted: lorawan.AES128Key{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
					MinMcFCnt:      100,
					MaxMcFCnt:      200,
				},
			},
			Bytes: []byte{0x2, 0x2, 0x4, 0x3, 0x2, 0x1, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x64, 0x0, 0x0, 0x0, 0xc8, 0x0, 0x0, 0x0},
		},
		{
			Name:                   "McGroupSetupReq invalid bytes",
			Bytes:                  []byte{0x02, 0x02, 0x04, 0x03, 0x02, 0x01, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01, 0x64, 0x00, 0x00, 0x00, 0xc8, 0x00, 0x00},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 29 bytes are expected"),
		},
		{
			Name:   "McGroupSetupAns",
			Uplink: true,
			Command: Command{
				CID: McGroupSetupAns,
				Payload: &McGroupSetupAnsPayload{
					McGroupIDHeader: McGroupSetupAnsPayloadMcGroupIDHeader{
						IDError:   true,
						McGroupID: 3,
					},
				},
			},
			Bytes: []byte{0x02, 0x07},
		},
		{
			Name:                   "McGroupSetupAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 1 bytes are expected"),
		},
		{
			Name: "McGroupDeleteReq",
			Command: Command{
				CID: McGroupDeleteReq,
				Payload: &McGroupDeleteReqPayload{
					McGroupIDHeader: McGroupDeleteReqPayloadMcGroupIDHeader{
						McGroupID: 3,
					},
				},
			},
			Bytes: []byte{0x03, 0x03},
		},
		{
			Name:   "McGroupDeleteAns",
			Uplink: true,
			Command: Command{
				CID: McGroupDeleteAns,
				Payload: &McGroupDeleteAnsPayload{
					McGroupIDHeader: McGroupDeleteAnsPayloadMcGroupIDHeader{
						McGroupUndefined: true,
						McGroupID:        1,
					},
				},
			},
			Bytes: []byte{0x03, 0x05},
		},
		{
			Name:                   "McGroupDeleteAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x03},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 1 bytes are expected"),
		},
		{
			Name: "McClassCSessionReq",
			Command: Command{
				CID: McClassCSessionReq,
				Payload: &McClassCSessionReqPayload{
					McGroupIDHeader: McClassCSessionReqPayloadMcGroupIDHeader{
						McGroupID: 3,
					},
					SessionTime: 134480385,
					SessionTimeOut: McClassCSessionReqPayloadSessionTimeOut{
						TimeOut: 8,
					},
					DLFrequency: 868100000,
					DR:          4,
				},
			},
			Bytes: []byte{0x04, 0x03, 0x01, 0x02, 0x04, 0x08, 0x08, 0x28, 0x76, 0x84, 0x04},
		},
		{
			Name:                   "McClassCSessionReq invalid bytes",
			Bytes:                  []byte{0x04, 0x03, 0x01, 0x02, 0x04, 0x08, 0x08, 0x28, 0x76, 0x84},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 10 bytes are expected"),
		},
		{
			Name:   "McClassCSessionAns no errors",
			Uplink: true,
			Command: Command{
				CID: McClassCSessionAns,
				Payload: &McClassCSessionAnsPayload{
					StatusAndMcGroupID: McClassCSessionAnsPayloadStatusAndMcGroupID{
						McGroupID: 3,
					},
					TimeToStart: &timeToStart,
				},
			},
			Bytes: []byte{0x04, 0x03, 0x01, 0x02, 0x04},
		},
		{
			Name:                   "McClassCSessionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x04, 0x03, 0x01, 0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 4 bytes are expected"),
		},
		{
			Name:   "McClassCSessionAns with errors",
			Uplink: true,
			Command: Command{
				CID: McClassCSessionAns,
				Payload: &McClassCSessionAnsPayload{
					StatusAndMcGroupID: McClassCSessionAnsPayloadStatusAndMcGroupID{
						McGroupID:        3,
						McGroupUndefined: true,
						FreqError:        true,
						DRError:          true,
					},
				},
			},
			Bytes: []byte{0x04, 0x1f},
		},
		{
			Name:   "McClassCSessionAns TimeToStart missing",
			Uplink: true,
			Command: Command{
				CID: McClassCSessionAns,
				Payload: &McClassCSessionAnsPayload{
					StatusAndMcGroupID: McClassCSessionAnsPayloadStatusAndMcGroupID{
						McGroupID: 3,
					},
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/multicastsetup: TimeToStart must not be nil"),
		},
		{
			Name:   "McClassCSessionAns TimeToStart not expected",
			Uplink: true,
			Command: Command{
				CID: McClassCSessionAns,
				Payload: &McClassCSessionAnsPayload{
					StatusAndMcGroupID: McClassCSessionAnsPayloadStatusAndMcGroupID{
						McGroupID:        3,
						McGroupUndefined: true,
						FreqError:        true,
						DRError:          true,
					},
					TimeToStart: &timeToStart,
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/multicastsetup: TimeToStart must be nil when StatusAndMcGroupID contains an error"),
		},
		{
			Name: "McClassBSessionReq",
			Command: Command{
				CID: McClassBSessionReq,
				Payload: &McClassBSessionReqPayload{
					McGroupIDHeader: McClassBSessionReqPayloadMcGroupIDHeader{
						McGroupID: 3,
					},
					SessionTime: 134480385,
					TimeOutPeriodicity: McClassBSessionReqPayloadTimeOutPeriodicity{
						TimeOut:     8,
						Periodicity: 4,
					},
					DLFrequency: 868100000,
					DR:          5,
				},
			},
			Bytes: []byte{0x05, 0x03, 0x01, 0x02, 0x04, 0x08, 0x48, 0x28, 0x76, 0x84, 0x05},
		},
		{
			Name:                   "McClassBSessionReq invalid bytes",
			Bytes:                  []byte{0x05, 0x03, 0x01, 0x02, 0x04, 0x08, 0x48, 0x28, 0x76, 0x84},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 10 bytes are expected"),
		},
		{
			Name:   "McClassBSessionAns no errors",
			Uplink: true,
			Command: Command{
				CID: McClassBSessionAns,
				Payload: &McClassBSessionAnsPayload{
					StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
						McGroupID: 3,
					},
					TimeToStart: &timeToStart,
				},
			},
			Bytes: []byte{0x05, 0x03, 0x01, 0x02, 0x04},
		},
		{
			Name:                   "McClassBSessionAns invalid bytes",
			Uplink:                 true,
			Bytes:                  []byte{0x05, 0x03, 0x01, 0x02},
			ExpectedUnmarshalError: errors.New("lorawan/applayer/multicastsetup: 4 bytes are expected"),
		},
		{
			Name:   "McClassBSessionAns with errors",
			Uplink: true,
			Command: Command{
				CID: McClassBSessionAns,
				Payload: &McClassBSessionAnsPayload{
					StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
						McGroupID:        3,
						McGroupUndefined: true,
						FreqError:        true,
						DRError:          true,
					},
				},
			},
			Bytes: []byte{0x05, 0x1f},
		},
		{
			Name:   "McClassBSessionAns TimeToStart missing",
			Uplink: true,
			Command: Command{
				CID: McClassBSessionAns,
				Payload: &McClassBSessionAnsPayload{
					StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
						McGroupID: 3,
					},
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/multicastsetup: TimeToStart must not be nil"),
		},
		{
			Name:   "McClassBSessionAns TimeToStart not expected",
			Uplink: true,
			Command: Command{
				CID: McClassBSessionAns,
				Payload: &McClassBSessionAnsPayload{
					StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
						McGroupID:        3,
						McGroupUndefined: true,
						FreqError:        true,
						DRError:          true,
					},
					TimeToStart: &timeToStart,
				},
			},
			ExpectedMarshalError: errors.New("lorawan/applayer/multicastsetup: TimeToStart must be nil when StatusAndMcGroupID contains an error"),
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

func TestUnmarshalCommands(t *testing.T) {
	assert := require.New(t)
	timeToStart := uint32(262657)

	// encode two variable sized commands and try to decode them

	commands := Commands{
		{
			CID: McClassBSessionAns,
			Payload: &McClassBSessionAnsPayload{
				StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
					McGroupID: 3,
				},
				TimeToStart: &timeToStart,
			},
		},
		Command{
			CID: McClassBSessionAns,
			Payload: &McClassBSessionAnsPayload{
				StatusAndMcGroupID: McClassBSessionAnsPayloadStatusAndMcGroupID{
					McGroupID:        3,
					McGroupUndefined: true,
					FreqError:        true,
					DRError:          true,
				},
			},
		},
	}

	b, err := commands.MarshalBinary()
	assert.NoError(err)

	assert.Equal(
		[]byte{0x05, 0x03, 0x01, 0x02, 0x04, 0x05, 0x1f},
		b,
	)

	var cmds Commands
	assert.NoError(cmds.UnmarshalBinary(true, b))
	assert.Equal(commands, cmds)
}
