package backend

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/band"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ClientTestSuite struct {
	suite.Suite

	client      Client
	server      *httptest.Server
	apiRequest  string
	apiResponse string
}

func (ts *ClientTestSuite) SetupSuite() {
	assert := require.New(ts.T())
	var err error
	ts.server = httptest.NewServer(http.HandlerFunc(ts.apiHandler))
	ts.client, err = NewClient("010101", "020202", ts.server.URL, "", "", "")
	assert.NoError(err)
}

func (ts *ClientTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *ClientTestSuite) TestPRStartReq() {
	assert := require.New(ts.T())

	devAddr := lorawan.DevAddr{1, 2, 3, 4}
	devEUI := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	dr := 2
	uFreq := 868.1
	gwCount := 1
	rssi := 10
	snr := 5.5

	req := PRStartReqPayload{
		BasePayload: BasePayload{
			ProtocolVersion: ProtocolVersion1_0,
			SenderID:        "010101",
			ReceiverID:      "020202",
			TransactionID:   123,
			MessageType:     PRStartReq,
		},
		PHYPayload: []byte{1, 2, 3, 4},
		ULMetaData: ULMetaData{
			DevAddr:  &devAddr,
			DataRate: &dr,
			ULFreq:   &uFreq,
			RecvTime: ISO8601Time(time.Now()),
			RFRegion: string(band.EU868),
			GWCnt:    &gwCount,
			GWInfo: []GWInfoElement{
				{
					ID:       []byte{1, 2, 3, 4},
					RFRegion: string(band.EU868),
					RSSI:     &rssi,
					SNR:      &snr,
				},
			},
		},
	}
	reqB, err := json.Marshal(req)
	assert.NoError(err)

	lifetime := 60

	resp := PRStartAnsPayload{
		BasePayloadResult: BasePayloadResult{
			BasePayload: BasePayload{
				ProtocolVersion: ProtocolVersion1_0,
				SenderID:        "020202",
				ReceiverID:      "010101",
				TransactionID:   123,
				MessageType:     PRStartAns,
			},
			Result: Result{
				ResultCode: Success,
			},
		},
		DevEUI:   &devEUI,
		Lifetime: &lifetime,
		NwkSKey: &KeyEnvelope{
			AESKey: HEXBytes{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
	}
	respB, err := json.Marshal(resp)
	assert.NoError(err)
	ts.apiResponse = string(respB)

	apiResp, err := ts.client.PRStartReq(context.Background(), req)
	assert.NoError(err)
	assert.Equal(resp, apiResp)

	assert.Equal(string(reqB), ts.apiRequest)
}

func (ts *ClientTestSuite) TestPRStopReq() {
	assert := require.New(ts.T())

	req := PRStopReqPayload{
		BasePayload: BasePayload{
			ProtocolVersion: ProtocolVersion1_0,
			SenderID:        "010101",
			ReceiverID:      "020202",
			TransactionID:   123,
			MessageType:     PRStopReq,
		},
		DevEUI: lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
	}
	reqB, err := json.Marshal(req)
	assert.NoError(err)

	resp := PRStopAnsPayload{
		BasePayloadResult: BasePayloadResult{
			BasePayload: BasePayload{
				ProtocolVersion: ProtocolVersion1_0,
				SenderID:        "020202",
				ReceiverID:      "010101",
				TransactionID:   123,
				MessageType:     PRStopAns,
			},
			Result: Result{
				ResultCode: Success,
			},
		},
	}
	respB, err := json.Marshal(resp)
	assert.NoError(err)
	ts.apiResponse = string(respB)

	apiResp, err := ts.client.PRStopReq(context.Background(), req)
	assert.NoError(err)
	assert.Equal(resp, apiResp)

	assert.Equal(string(reqB), ts.apiRequest)
}

func (ts *ClientTestSuite) TestXmitDataReq() {
	assert := require.New(ts.T())

	devAddr := lorawan.DevAddr{1, 2, 3, 4}
	devEUI := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	dr := 2
	uFreq := 868.1
	gwCount := 1
	rssi := 10
	snr := 5.5

	req := XmitDataReqPayload{
		BasePayload: BasePayload{
			ProtocolVersion: ProtocolVersion1_0,
			SenderID:        "010101",
			ReceiverID:      "020202",
			TransactionID:   123,
			MessageType:     XmitDataReq,
		},
		PHYPayload: []byte{1, 2, 3},
		ULMetaData: &ULMetaData{
			DevAddr:  &devAddr,
			DevEUI:   &devEUI,
			DataRate: &dr,
			ULFreq:   &uFreq,
			RecvTime: ISO8601Time(time.Now()),
			RFRegion: string(band.EU868),
			GWCnt:    &gwCount,
			GWInfo: []GWInfoElement{
				{
					ID:       []byte{1, 2, 3, 4},
					RFRegion: string(band.EU868),
					RSSI:     &rssi,
					SNR:      &snr,
				},
			},
		},
	}
	reqB, err := json.Marshal(req)
	assert.NoError(err)

	resp := XmitDataAnsPayload{
		BasePayloadResult: BasePayloadResult{
			BasePayload: BasePayload{
				ProtocolVersion: ProtocolVersion1_0,
				SenderID:        "020202",
				ReceiverID:      "010101",
				TransactionID:   123,
				MessageType:     XmitDataAns,
			},
			Result: Result{
				ResultCode: Success,
			},
		},
	}
	respB, err := json.Marshal(resp)
	assert.NoError(err)
	ts.apiResponse = string(respB)

	apiResp, err := ts.client.XmitDataReq(context.Background(), req)
	assert.NoError(err)
	assert.Equal(resp, apiResp)

	assert.Equal(string(reqB), ts.apiRequest)
}

func (ts *ClientTestSuite) TestProfileReq() {
	assert := require.New(ts.T())

	devEUI := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	timestamp := ISO8601Time(time.Now().UTC().Truncate(time.Second))
	handover := Handover

	req := ProfileReqPayload{
		BasePayload: BasePayload{
			ProtocolVersion: ProtocolVersion1_0,
			SenderID:        "010101",
			ReceiverID:      "020202",
			TransactionID:   123,
			MessageType:     ProfileReq,
		},
		DevEUI: devEUI,
	}
	reqB, err := json.Marshal(req)
	assert.NoError(err)

	resp := ProfileAnsPayload{
		BasePayloadResult: BasePayloadResult{
			BasePayload: BasePayload{
				ProtocolVersion: ProtocolVersion1_0,
				SenderID:        "020202",
				ReceiverID:      "010101",
				TransactionID:   123,
				MessageType:     ProfileAns,
			},
			Result: Result{
				ResultCode: Success,
			},
		},
		DeviceProfile: &DeviceProfile{
			DeviceProfileID: "test-1234",
			MACVersion:      "1.0.3",
			SupportsJoin:    true,
		},
		DeviceProfileTimestamp: &timestamp,
		RoamingActivationType:  &handover,
	}
	respB, err := json.Marshal(resp)
	assert.NoError(err)
	ts.apiResponse = string(respB)

	apiResp, err := ts.client.ProfileReq(context.Background(), req)
	assert.NoError(err)
	assert.Equal(resp, apiResp)

	assert.Equal(string(reqB), ts.apiRequest)
}

func (ts *ClientTestSuite) TestHomeNSReq() {
	assert := require.New(ts.T())

	devEUI := lorawan.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	netID := lorawan.NetID{1, 2, 3}

	req := HomeNSReqPayload{
		BasePayload: BasePayload{
			ProtocolVersion: ProtocolVersion1_0,
			SenderID:        "010101",
			ReceiverID:      "020202",
			TransactionID:   123,
			MessageType:     HomeNSReq,
		},
		DevEUI: devEUI,
	}
	reqB, err := json.Marshal(req)
	assert.NoError(err)

	resp := HomeNSAnsPayload{
		BasePayloadResult: BasePayloadResult{
			BasePayload: BasePayload{
				ProtocolVersion: ProtocolVersion1_0,
				SenderID:        "020202",
				ReceiverID:      "010101",
				TransactionID:   123,
				MessageType:     HomeNSAns,
			},
			Result: Result{
				ResultCode: Success,
			},
		},
		HNetID: netID,
	}
	respB, err := json.Marshal(resp)
	assert.NoError(err)
	ts.apiResponse = string(respB)

	apiResp, err := ts.client.HomeNSReq(context.Background(), req)
	assert.NoError(err)
	assert.Equal(resp, apiResp)

	assert.Equal(string(reqB), ts.apiRequest)
}

func (ts *ClientTestSuite) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	ts.apiRequest = string(b)
	w.Write([]byte(ts.apiResponse))
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
