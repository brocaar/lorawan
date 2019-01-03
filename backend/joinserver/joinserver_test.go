package joinserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

type JoinServerTestSuite struct {
	suite.Suite
	deviceKeys map[lorawan.EUI64]DeviceKeys
	asKEKLabel string
	keks       map[string][]byte

	server *httptest.Server
}

func (ts *JoinServerTestSuite) SetupSuite() {
	assert := require.New(ts.T())

	ts.deviceKeys = make(map[lorawan.EUI64]DeviceKeys)
	ts.keks = make(map[string][]byte)

	config := HandlerConfig{
		GetDeviceKeysByDevEUIFunc: ts.getDeviceKeys,
		GetASKEKLabelByDevEUIFunc: ts.getASKEKLabelByDevEUI,
		GetKEKByLabelFunc:         ts.getKEKByLabel,
	}

	handler, err := NewHandler(config)
	assert.NoError(err)

	ts.server = httptest.NewServer(handler)
}

func (ts *JoinServerTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *JoinServerTestSuite) getDeviceKeys(devEUI lorawan.EUI64) (DeviceKeys, error) {
	if dk, ok := ts.deviceKeys[devEUI]; ok {
		return dk, nil
	}

	return DeviceKeys{}, ErrDevEUINotFound
}

func (ts *JoinServerTestSuite) getASKEKLabelByDevEUI(devEUI lorawan.EUI64) (string, error) {
	return ts.asKEKLabel, nil
}

func (ts *JoinServerTestSuite) getKEKByLabel(label string) ([]byte, error) {
	return ts.keks[label], nil
}

func (ts *JoinServerTestSuite) TestJoinRequest() {
	assert := require.New(ts.T())

	cFList := lorawan.CFList{
		CFListType: lorawan.CFListChannel,
		Payload: &lorawan.CFListChannelPayload{
			Channels: [5]uint32{
				868700000,
				868900000,
			},
		},
	}
	cFListB, err := cFList.MarshalBinary()
	assert.NoError(err)

	dk := DeviceKeys{
		DevEUI:    lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		NwkKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		JoinNonce: 65536,
	}

	validJRPHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			DevEUI:   dk.DevEUI,
			JoinEUI:  lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: 258,
		},
	}
	assert.NoError(validJRPHY.SetUplinkJoinMIC(dk.NwkKey))
	validJRPHYBytes, err := validJRPHY.MarshalBinary()
	assert.NoError(err)

	invalidMICJRPHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinRequestPayload{
			DevEUI:   dk.DevEUI,
			JoinEUI:  lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
			DevNonce: 258,
		},
	}
	assert.NoError(invalidMICJRPHY.SetUplinkJoinMIC(lorawan.AES128Key{}))
	invalidMICJRPHYBytes, err := invalidMICJRPHY.MarshalBinary()
	assert.NoError(err)

	validJAPHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce: 65536,
			HomeNetID: lorawan.NetID{1, 2, 3},
			DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				RX2DataRate: 5,
				RX1DROffset: 1,
			},
			RXDelay: 1,
			CFList: &lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						868700000,
						868900000,
					},
				},
			},
		},
	}
	assert.NoError(validJAPHY.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, dk.NwkKey))
	assert.NoError(validJAPHY.EncryptJoinAcceptPayload(dk.NwkKey))
	validJAPHYBytes, err := validJAPHY.MarshalBinary()
	assert.NoError(err)

	validJAPHYLW11 := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce: 65536,
			HomeNetID: lorawan.NetID{1, 2, 3},
			DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				OptNeg:      true,
				RX2DataRate: 5,
				RX1DROffset: 1,
			},
			RXDelay: 1,
			CFList: &lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						868700000,
						868900000,
					},
				},
			},
		},
	}
	jsIntKey, err := getJSIntKey(dk.NwkKey, dk.DevEUI)
	assert.NoError(err)
	assert.NoError(validJAPHYLW11.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 258, jsIntKey))

	assert.NoError(validJAPHYLW11.EncryptJoinAcceptPayload(dk.NwkKey))
	validJAPHYLW11Bytes, err := validJAPHYLW11.MarshalBinary()
	assert.NoError(err)

	tests := []struct {
		Name               string
		DeviceKeys         map[lorawan.EUI64]DeviceKeys
		RequestPayload     backend.JoinReqPayload
		ExpectedAnsPayload backend.JoinAnsPayload
		ASKEKLabel         string
		KEKs               map[string][]byte
	}{
		{
			Name:       "valid join-request (LoRaWAN 1.0)",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.JoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.JoinReq,
				},
				MACVersion: "1.0.2",
				PHYPayload: backend.HEXBytes(validJRPHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.JoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.JoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(validJAPHYBytes),
				NwkSKey: &backend.KeyEnvelope{
					AESKey: []byte{223, 83, 195, 95, 48, 52, 204, 206, 208, 255, 53, 76, 112, 222, 4, 223},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{146, 123, 156, 145, 17, 131, 207, 254, 76, 178, 255, 75, 117, 84, 95, 109},
				},
			},
		},
		{
			Name:       "valid join-request (LoRaWAN 1.1)",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.JoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.JoinReq,
				},
				MACVersion: "1.1.0",
				PHYPayload: backend.HEXBytes(validJRPHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					OptNeg:      true,
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.JoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.JoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(validJAPHYLW11Bytes),
				FNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{83, 127, 138, 174, 137, 108, 121, 224, 21, 209, 2, 208, 98, 134, 53, 78},
				},
				SNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{88, 148, 152, 153, 48, 146, 207, 219, 95, 210, 224, 42, 199, 81, 11, 241},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					AESKey: []byte{152, 152, 40, 60, 79, 102, 235, 108, 111, 213, 22, 88, 130, 4, 108, 64},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{1, 98, 18, 21, 209, 202, 8, 254, 191, 12, 96, 44, 194, 173, 144, 250},
				},
			},
		},
		{
			Name:       "valid join-request (LoRaWAN 1.1) with KEK",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			ASKEKLabel: "lora-app-server",
			KEKs: map[string][]byte{
				"010203":          make([]byte, 16),
				"lora-app-server": make([]byte, 16),
			},
			RequestPayload: backend.JoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.JoinReq,
				},
				MACVersion: "1.1.0",
				PHYPayload: backend.HEXBytes(validJRPHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					OptNeg:      true,
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.JoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.JoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(validJAPHYLW11Bytes),
				FNwkSIntKey: &backend.KeyEnvelope{
					KEKLabel: "010203",
					AESKey:   []byte{87, 85, 230, 195, 36, 30, 231, 230, 100, 111, 15, 254, 135, 120, 122, 0, 44, 249, 228, 176, 131, 73, 143, 0},
				},
				SNwkSIntKey: &backend.KeyEnvelope{
					KEKLabel: "010203",
					AESKey:   []byte{246, 176, 184, 31, 61, 48, 41, 18, 85, 145, 192, 176, 184, 141, 118, 201, 59, 72, 172, 164, 4, 22, 133, 211},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					KEKLabel: "010203",
					AESKey:   []byte{78, 225, 236, 219, 189, 151, 82, 239, 109, 226, 140, 65, 233, 189, 174, 37, 39, 206, 241, 242, 2, 127, 157, 247},
				},
				AppSKey: &backend.KeyEnvelope{
					KEKLabel: "lora-app-server",
					AESKey:   []byte{248, 215, 201, 250, 55, 176, 209, 198, 53, 78, 109, 184, 225, 157, 157, 122, 180, 229, 199, 88, 30, 159, 30, 32},
				},
			},
		},
		{
			Name:       "join-request with invalid mic",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.JoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.JoinReq,
				},
				MACVersion: "1.0.2",
				PHYPayload: backend.HEXBytes(invalidMICJRPHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.JoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.JoinAns,
				},
				Result: backend.Result{
					ResultCode:  backend.MICFailed,
					Description: "invalid mic",
				},
			},
		},
		{
			Name: "join-request for unknown device",
			RequestPayload: backend.JoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.JoinReq,
				},
				MACVersion: "1.0.2",
				PHYPayload: backend.HEXBytes(validJRPHYBytes),
				DevEUI:     lorawan.EUI64{1, 1, 1, 1, 1, 1, 1, 1},
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.JoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.JoinAns,
				},
				Result: backend.Result{
					ResultCode:  backend.UnknownDevEUI,
					Description: ErrDevEUINotFound.Error(),
				},
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			ts.deviceKeys = tst.DeviceKeys
			ts.asKEKLabel = tst.ASKEKLabel
			ts.keks = tst.KEKs

			b, err := json.Marshal(tst.RequestPayload)
			assert.NoError(err)

			resp, err := http.Post(ts.server.URL, "application/json", bytes.NewReader(b))
			assert.NoError(err)
			defer resp.Body.Close()

			var ansPayload backend.JoinAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&ansPayload))

			assert.Equal(tst.ExpectedAnsPayload, ansPayload)
		})
	}
}

func (ts *JoinServerTestSuite) TestRejoinRequest() {
	assert := require.New(ts.T())

	cFList := lorawan.CFList{
		CFListType: lorawan.CFListChannel,
		Payload: &lorawan.CFListChannelPayload{
			Channels: [5]uint32{
				868700000,
				868900000,
			},
		},
	}
	cFListB, err := cFList.MarshalBinary()
	assert.NoError(err)

	dk := DeviceKeys{
		DevEUI:    lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
		NwkKey:    lorawan.AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		JoinNonce: 65536,
	}

	jsIntKey, err := getJSIntKey(dk.NwkKey, dk.DevEUI)
	assert.NoError(err)
	jsEncKey, err := getJSEncKey(dk.NwkKey, dk.DevEUI)
	assert.NoError(err)

	rj0PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.RejoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.RejoinRequestType02Payload{
			RejoinType: lorawan.RejoinRequestType0,
			NetID:      lorawan.NetID{1, 2, 3},
			DevEUI:     dk.DevEUI,
			RJCount0:   123,
		},
	}
	// no need to set the MIC as it is not validated by the js
	rj0PHYBytes, err := rj0PHY.MarshalBinary()
	assert.NoError(err)

	rj1PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.RejoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.RejoinRequestType1Payload{
			RejoinType: lorawan.RejoinRequestType1,
			JoinEUI:    lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1},
			DevEUI:     dk.DevEUI,
			RJCount1:   123,
		},
	}
	assert.NoError(rj1PHY.SetUplinkJoinMIC(jsIntKey))
	rj1PHYBytes, err := rj1PHY.MarshalBinary()
	assert.NoError(err)

	rj2PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.RejoinRequest,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.RejoinRequestType02Payload{
			RejoinType: lorawan.RejoinRequestType2,
			NetID:      lorawan.NetID{1, 2, 3},
			DevEUI:     dk.DevEUI,
			RJCount0:   123,
		},
	}
	// no need to set the MIC as it is not validated by the js
	rj2PHYBytes, err := rj2PHY.MarshalBinary()
	assert.NoError(err)

	ja0PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(dk.JoinNonce),
			HomeNetID: lorawan.NetID{1, 2, 3},
			DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				OptNeg:      true,
				RX2DataRate: 5,
				RX1DROffset: 1,
			},
			RXDelay: 1,
			CFList: &lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						868700000,
						868900000,
					},
				},
			},
		},
	}
	assert.NoError(ja0PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType0, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey))

	assert.NoError(ja0PHY.EncryptJoinAcceptPayload(jsEncKey))
	ja0PHYBytes, err := ja0PHY.MarshalBinary()
	assert.NoError(err)

	ja1PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(dk.JoinNonce),
			HomeNetID: lorawan.NetID{1, 2, 3},
			DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				OptNeg:      true,
				RX2DataRate: 5,
				RX1DROffset: 1,
			},
			RXDelay: 1,
			CFList: &lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						868700000,
						868900000,
					},
				},
			},
		},
	}
	assert.NoError(ja1PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType1, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey))
	assert.NoError(ja1PHY.EncryptJoinAcceptPayload(jsEncKey))
	ja1PHYBytes, err := ja1PHY.MarshalBinary()
	assert.NoError(err)

	ja2PHY := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce: lorawan.JoinNonce(dk.JoinNonce),
			HomeNetID: lorawan.NetID{1, 2, 3},
			DevAddr:   lorawan.DevAddr{1, 2, 3, 4},
			DLSettings: lorawan.DLSettings{
				OptNeg:      true,
				RX2DataRate: 5,
				RX1DROffset: 1,
			},
			RXDelay: 1,
		},
	}
	assert.NoError(ja2PHY.SetDownlinkJoinMIC(lorawan.RejoinRequestType2, lorawan.EUI64{8, 7, 6, 5, 4, 3, 2, 1}, 123, jsIntKey))
	assert.NoError(ja2PHY.EncryptJoinAcceptPayload(jsEncKey))
	ja2PHYBytes, err := ja2PHY.MarshalBinary()
	assert.NoError(err)

	tests := []struct {
		Name               string
		DeviceKeys         map[lorawan.EUI64]DeviceKeys
		RequestPayload     backend.RejoinReqPayload
		ExpectedAnsPayload backend.RejoinAnsPayload
		ASKEKLabel         string
		KEKs               map[string][]byte
	}{
		{
			Name:       "valid rejoin-request type 0",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.RejoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.RejoinReq,
				},
				MACVersion: "1.1.0",
				PHYPayload: backend.HEXBytes(rj0PHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					OptNeg:      true,
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.RejoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.RejoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(ja0PHYBytes),
				SNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
				},
				FNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					AESKey: []byte{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
			},
		},
		{
			Name:       "valid rejoin-request type 1",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.RejoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.RejoinReq,
				},
				MACVersion: "1.1.0",
				PHYPayload: backend.HEXBytes(rj1PHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					OptNeg:      true,
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
				CFList:  backend.HEXBytes(cFListB),
			},
			ExpectedAnsPayload: backend.RejoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.RejoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(ja1PHYBytes),
				SNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
				},
				FNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					AESKey: []byte{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
			},
		},
		{
			Name:       "valid rejoin-request type 2",
			DeviceKeys: map[lorawan.EUI64]DeviceKeys{dk.DevEUI: dk},
			RequestPayload: backend.RejoinReqPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "010203",
					ReceiverID:      "0807060504030201",
					TransactionID:   1234,
					MessageType:     backend.RejoinReq,
				},
				MACVersion: "1.1.0",
				PHYPayload: backend.HEXBytes(rj2PHYBytes),
				DevEUI:     dk.DevEUI,
				DevAddr:    lorawan.DevAddr{1, 2, 3, 4},
				DLSettings: lorawan.DLSettings{
					OptNeg:      true,
					RX2DataRate: 5,
					RX1DROffset: 1,
				},
				RxDelay: 1,
			},
			ExpectedAnsPayload: backend.RejoinAnsPayload{
				BasePayload: backend.BasePayload{
					ProtocolVersion: backend.ProtocolVersion1_0,
					SenderID:        "0807060504030201",
					ReceiverID:      "010203",
					TransactionID:   1234,
					MessageType:     backend.RejoinAns,
				},
				Result: backend.Result{
					ResultCode: backend.Success,
				},
				PHYPayload: backend.HEXBytes(ja2PHYBytes),
				SNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{84, 115, 118, 176, 7, 14, 169, 150, 78, 61, 226, 98, 252, 231, 85, 145},
				},
				FNwkSIntKey: &backend.KeyEnvelope{
					AESKey: []byte{15, 235, 84, 189, 47, 133, 75, 254, 195, 103, 254, 91, 27, 132, 16, 55},
				},
				NwkSEncKey: &backend.KeyEnvelope{
					AESKey: []byte{212, 9, 208, 87, 17, 14, 159, 221, 5, 199, 126, 12, 85, 63, 119, 244},
				},
				AppSKey: &backend.KeyEnvelope{
					AESKey: []byte{11, 25, 22, 151, 83, 252, 60, 31, 222, 161, 118, 106, 12, 34, 117, 225},
				},
			},
		},
	}

	for _, tst := range tests {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			ts.deviceKeys = tst.DeviceKeys
			ts.asKEKLabel = tst.ASKEKLabel
			ts.keks = tst.KEKs

			b, err := json.Marshal(tst.RequestPayload)
			assert.NoError(err)

			resp, err := http.Post(ts.server.URL, "application/json", bytes.NewReader(b))
			assert.NoError(err)
			defer resp.Body.Close()

			var ansPayload backend.RejoinAnsPayload
			assert.NoError(json.NewDecoder(resp.Body).Decode(&ansPayload))

			assert.Equal(tst.ExpectedAnsPayload, ansPayload)
		})
	}
}

func TestJoinServer(t *testing.T) {
	suite.Run(t, new(JoinServerTestSuite))
}
