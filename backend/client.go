package backend

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

// Errors.
var (
	ErrAsyncTimeout = errors.New("async timeout")
)

// Client defines the backend client interface.
type Client interface {
	// GetSenderID returns the SenderID.
	GetSenderID() string
	// GetReceiverID returns the ReceiverID.
	GetReceiverID() string
	// IsAsync returns a bool indicating if the client is async.
	IsAsync() bool
	// GetRandomTransactionID returns a random transaction id.
	GetRandomTransactionID() uint32
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
	// SendAnswer sends the async answer.
	SendAnswer(context.Context, Answer) error
	// HandleAnswer handles an async answer.
	HandleAnswer(context.Context, Answer) error
}

// ClientConfig holds the backend client configuration.
type ClientConfig struct {
	SenderID   string
	ReceiverID string
	Server     string
	CACert     string
	TLSCert    string
	TLSKey     string

	// RedisClient holds the optional Redis database client. When set the client
	// will use the aysnc protocol scheme. In this case the client will wait
	// AsyncTimeout before returning a timeout error.
	RedisClient redis.UniversalClient

	// AsyncTimeout defines the async timeout. This must be set when RedisClient
	// is set.
	AsyncTimeout time.Duration
}

// NewClient creates a new Client.
func NewClient(config ClientConfig) (Client, error) {
	httpClient := http.DefaultClient

	if config.CACert != "" || config.TLSCert != "" || config.TLSKey != "" {
		tlsConfig := &tls.Config{}

		if config.CACert != "" {
			rawCACert, err := ioutil.ReadFile(config.CACert)
			if err != nil {
				return nil, errors.Wrap(err, "read ca cert error")
			}

			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(rawCACert) {
				return nil, errors.New("append ca cert to pool error")
			}

			tlsConfig.RootCAs = caCertPool
		}

		if config.TLSCert != "" || config.TLSKey != "" {
			cert, err := tls.LoadX509KeyPair(config.TLSCert, config.TLSKey)
			if err != nil {
				return nil, errors.Wrap(err, "load x509 keypair error")
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	}

	return &client{
		server:          config.Server,
		httpClient:      httpClient,
		senderID:        config.SenderID,
		receiverID:      config.ReceiverID,
		protocolVersion: ProtocolVersion1_0,
		redisClient:     config.RedisClient,
		asyncTimeout:    config.AsyncTimeout,
	}, nil

}

type client struct {
	server          string
	httpClient      *http.Client
	protocolVersion string
	senderID        string
	receiverID      string
	redisClient     redis.UniversalClient
	asyncTimeout    time.Duration
}

func (c *client) GetSenderID() string {
	return c.senderID
}

func (c *client) GetReceiverID() string {
	return c.receiverID
}

func (c *client) IsAsync() bool {
	return c.redisClient != nil
}

func (c *client) PRStartReq(ctx context.Context, pl PRStartReqPayload) (PRStartAnsPayload, error) {
	pl.BasePayload.ProtocolVersion = c.protocolVersion
	pl.BasePayload.SenderID = c.senderID
	pl.BasePayload.ReceiverID = c.receiverID
	pl.BasePayload.MessageType = PRStartReq
	if pl.BasePayload.TransactionID == 0 {
		pl.BasePayload.TransactionID = c.GetRandomTransactionID()
	}

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
	if pl.BasePayload.TransactionID == 0 {
		pl.BasePayload.TransactionID = c.GetRandomTransactionID()
	}

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
	if pl.BasePayload.TransactionID == 0 {
		pl.BasePayload.TransactionID = c.GetRandomTransactionID()
	}

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
	if pl.BasePayload.TransactionID == 0 {
		pl.BasePayload.TransactionID = c.GetRandomTransactionID()
	}

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
	if pl.BasePayload.TransactionID == 0 {
		pl.BasePayload.TransactionID = c.GetRandomTransactionID()
	}

	var ans HomeNSAnsPayload

	if err := c.request(ctx, pl, &ans); err != nil {
		return ans, err
	}

	if ans.Result.ResultCode != Success {
		return ans, fmt.Errorf("response error, code: %s, description: %s", ans.Result.ResultCode, ans.Result.Description)
	}

	return ans, nil
}

func (c *client) request(ctx context.Context, pl Request, ans interface{}) error {
	b, err := json.Marshal(pl)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}

	responseChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// Setup async subscriber to receive response. Please note that we have to do
	// this before making the request, as the response might come in, before the
	// request has returned.
	if c.IsAsync() {
		key := c.getAsyncKey(pl.GetBasePayload().TransactionID)

		go func() {
			bb, err := c.readAsync(ctx, key)
			if err != nil {
				errorChan <- err
			} else {
				responseChan <- bb
			}
		}()
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.server, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "new request error")
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http post error")
	}
	defer resp.Body.Close()

	// If async is not used, the http response contains the API response payload.
	if !c.IsAsync() {
		bb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			errorChan <- err
		} else {
			responseChan <- bb
		}
	}

	select {
	case err := <-errorChan:
		return err
	case bb := <-responseChan:
		if err := json.Unmarshal(bb, ans); err != nil {
			return errors.Wrap(err, "unmarshal response error")
		}
	}

	return nil
}

func (c *client) HandleAnswer(ctx context.Context, pl Answer) error {
	if !c.IsAsync() {
		return errors.New("async is not configured")
	}

	b, err := json.Marshal(pl)
	if err != nil {
		return errors.Wrap(err, "marshal answer error")
	}

	err = c.redisClient.Publish(c.getAsyncKey(pl.GetBasePayload().TransactionID), b).Err()
	if err != nil {
		return errors.Wrap(err, "publish answer error")
	}

	return nil
}

func (c *client) SendAnswer(ctx context.Context, pl Answer) error {
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

	if resp.StatusCode != 200 {
		bb, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read body error")
		}
		return fmt.Errorf("expected: 200, got: %d (%s)", resp.StatusCode, string(bb))
	}

	return nil
}

func (c *client) GetRandomTransactionID() uint32 {
	b := make([]byte, 4)
	rand.Read(b)
	return binary.LittleEndian.Uint32(b)
}

func (c *client) getAsyncKey(id uint32) string {
	return fmt.Sprintf("lora:backend:async:%d", id)
}

func (c *client) readAsync(ctx context.Context, key string) ([]byte, error) {
	sub := c.redisClient.Subscribe(key)
	defer sub.Close()

	ch := sub.Channel()

	select {
	case msg := <-ch:
		return []byte(msg.Payload), nil
	case <-time.After(c.asyncTimeout):
		return nil, ErrAsyncTimeout
	}
}
