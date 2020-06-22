package backend

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// Client defines the backend client interface.
type Client interface {
	// PRStartReq method.
	PRStartReq(context.Context, PRStartReqPayload) (PRStartAnsPayload, error)
	// PRStopReq method.
	PRStopReq(context.Context, PRStopReqPayload) (PRStopAnsPayload, error)
	// XmitDataReq method.
	XmitDataReq(context.Context, XmitDataReqPayload) (XmitDataAnsPayload, error)
	// ProfileReq method.
	ProfileReq(context.Context, ProfileReqPayload) (ProfileAnsPayload, error)
	// HomeNSReq method.
	HomeNSReq(context.Context, HomeNSReqPayload) (HomeNSAnsPayload, error)
}

// NewClient creates a new Client.
func NewClient(senderID, receiverID string, server, caCert, tlsCert, tlsKey string) (Client, error) {
	if caCert == "" && tlsCert == "" && tlsKey == "" {
		return &client{
			server:          server,
			httpClient:      http.DefaultClient,
			senderID:        senderID,
			receiverID:      receiverID,
			protocolVersion: ProtocolVersion1_0,
		}, nil
	}

	tlsConfig := &tls.Config{}

	if caCert != "" {
		rawCACert, err := ioutil.ReadFile(caCert)
		if err != nil {
			return nil, errors.Wrap(err, "read ca cert error")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(rawCACert) {
			return nil, errors.New("append ca cert to pool error")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if tlsCert != "" || tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return nil, errors.Wrap(err, "load x509 keypair error")
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return &client{
		server: server,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}, nil

}

type client struct {
	server          string
	httpClient      *http.Client
	protocolVersion string
	senderID        string
	receiverID      string
}

func (c *client) PRStartReq(ctx context.Context, pl PRStartReqPayload) (PRStartAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = PRStartReq

	var ans PRStartAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) PRStopReq(ctx context.Context, pl PRStopReqPayload) (PRStopAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = PRStopReq

	var ans PRStopAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) XmitDataReq(ctx context.Context, pl XmitDataReqPayload) (XmitDataAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = XmitDataReq

	var ans XmitDataAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) ProfileReq(ctx context.Context, pl ProfileReqPayload) (ProfileAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = ProfileReq

	var ans ProfileAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) HomeNSReq(ctx context.Context, pl HomeNSReqPayload) (HomeNSAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = HomeNSReq

	var ans HomeNSAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) request(ctx context.Context, pl interface{}, ans interface{}) error {
	b, err := json.Marshal(pl)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	// TODO add context for cancellation
	resp, err := c.httpClient.Post(c.server, "application/json", bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "http post error")
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(ans)
	if err != nil {
		return errors.Wrap(err, "unmarshal response error")
	}

	return nil
}
