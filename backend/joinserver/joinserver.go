// Package joinserver provides a http.Handler interface which implements the
// join-server API as speficied by the LoRaWAN Backend Interfaces.
package joinserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

// DeviceKeys holds the device (root) keys and the join-nonce to be used
// for join-request and join-accepts.
// Note: it follows the LoRaWAN 1.1 key naming!
type DeviceKeys struct {
	DevEUI    lorawan.EUI64
	NwkKey    lorawan.AES128Key
	AppKey    lorawan.AES128Key
	JoinNonce int // the join-nonce that must be used for the join-accept
}

// HandlerConfig holds the join-server handler configuration.
type HandlerConfig struct {
	Logger                    *log.Logger
	GetDeviceKeysByDevEUIFunc func(devEUI lorawan.EUI64) (DeviceKeys, error) // ErrDevEUINotFound must be returned when the device does not exist
	GetKEKByLabelFunc         func(label string) ([]byte, error)             // must return an empty slice when no KEK exists for the given label
	GetASKEKLabelByDevEUIFunc func(devEUI lorawan.EUI64) (string, error)     // must return an empty string when no label exists
}

type handler struct {
	config HandlerConfig
	log    *log.Logger
}

// NewHandler creates a new join-sever handler.
func NewHandler(config HandlerConfig) (http.Handler, error) {
	if config.GetDeviceKeysByDevEUIFunc == nil {
		return nil, errors.New("backend/joinserver: GetDeviceKeysFunc must not be nil")
	}

	h := handler{
		config: config,
		log:    config.Logger,
	}

	if h.log == nil {
		h.log = &log.Logger{
			Out: ioutil.Discard,
		}
	}

	if h.config.GetKEKByLabelFunc == nil {
		h.log.Warning("backend/joinserver: get kek by label function is not set")

		h.config.GetKEKByLabelFunc = func(label string) ([]byte, error) {
			return nil, nil
		}
	}

	if h.config.GetASKEKLabelByDevEUIFunc == nil {
		h.log.Warning("backend/joinserver: get application-server kek by deveui function is not set")

		h.config.GetASKEKLabelByDevEUIFunc = func(devEUI lorawan.EUI64) (string, error) {
			return "", nil
		}
	}

	return &h, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var basePL backend.BasePayload

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		h.returnError(w, http.StatusInternalServerError, backend.Other, "read body error")
		return
	}

	err = json.Unmarshal(b, &basePL)
	if err != nil {
		h.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	h.log.WithFields(log.Fields{
		"message_type":   basePL.MessageType,
		"sender_id":      basePL.SenderID,
		"receiver_id":    basePL.ReceiverID,
		"transaction_id": basePL.TransactionID,
	}).Info("backend/joinserver: request received")

	switch basePL.MessageType {
	case backend.JoinReq:
		h.handleJoinReq(w, b)
	case backend.RejoinReq:
		h.handleRejoinReq(w, b)
	default:
		h.returnError(w, http.StatusBadRequest, backend.Other, fmt.Sprintf("invalid MessageType: %s", basePL.MessageType))
	}
}

func (h *handler) returnError(w http.ResponseWriter, code int, resultCode backend.ResultCode, msg string) {
	h.log.WithFields(log.Fields{
		"error": msg,
	}).Error("backend/joinserver: error handling request")

	w.WriteHeader(code)

	pl := backend.Result{
		ResultCode:  resultCode,
		Description: msg,
	}
	b, err := json.Marshal(pl)
	if err != nil {
		h.log.WithError(err).Error("backend/joinserver: marshal json error")
		return
	}

	w.Write(b)
}

func (h *handler) returnJoinReqError(w http.ResponseWriter, basePL backend.BasePayload, code int, resultCode backend.ResultCode, msg string) {
	jaPL := backend.JoinAnsPayload{
		BasePayload: backend.BasePayload{
			ProtocolVersion: backend.ProtocolVersion1_0,
			SenderID:        basePL.ReceiverID,
			ReceiverID:      basePL.SenderID,
			TransactionID:   basePL.TransactionID,
			MessageType:     backend.JoinAns,
		},
		Result: backend.Result{
			ResultCode:  resultCode,
			Description: msg,
		},
	}

	h.returnPayload(w, code, jaPL)
}

func (h *handler) returnRejoinReqError(w http.ResponseWriter, basePL backend.BasePayload, code int, resultCode backend.ResultCode, msg string) {
	jaPL := backend.RejoinAnsPayload{
		BasePayload: backend.BasePayload{
			ProtocolVersion: backend.ProtocolVersion1_0,
			SenderID:        basePL.ReceiverID,
			ReceiverID:      basePL.SenderID,
			TransactionID:   basePL.TransactionID,
			MessageType:     backend.RejoinAns,
		},
		Result: backend.Result{
			ResultCode:  resultCode,
			Description: msg,
		},
	}

	h.returnPayload(w, code, jaPL)
}

func (h *handler) returnPayload(w http.ResponseWriter, code int, pl interface{}) {
	w.WriteHeader(code)

	b, err := json.Marshal(pl)
	if err != nil {
		h.log.WithError(err).Error("backend/joinserver: marshal json error")
		return
	}

	w.Write(b)
}

func (h *handler) handleJoinReq(w http.ResponseWriter, b []byte) {
	var joinReqPL backend.JoinReqPayload
	err := json.Unmarshal(b, &joinReqPL)
	if err != nil {
		h.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	dk, err := h.config.GetDeviceKeysByDevEUIFunc(joinReqPL.DevEUI)
	if err != nil {
		switch err {
		case ErrDevEUINotFound:
			h.returnJoinReqError(w, joinReqPL.BasePayload, http.StatusBadRequest, backend.UnknownDevEUI, err.Error())
		default:
			h.returnJoinReqError(w, joinReqPL.BasePayload, http.StatusBadRequest, backend.Other, err.Error())
		}
		return
	}

	nsKEK, err := h.config.GetKEKByLabelFunc(joinReqPL.SenderID)
	if err != nil {
		h.returnJoinReqError(w, joinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	asKEKLabel, err := h.config.GetASKEKLabelByDevEUIFunc(joinReqPL.DevEUI)
	if err != nil {
		h.returnJoinReqError(w, joinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	asKEK, err := h.config.GetKEKByLabelFunc(asKEKLabel)
	if err != nil {
		h.returnJoinReqError(w, joinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	ans := handleJoinRequestWrapper(joinReqPL, dk, asKEKLabel, asKEK, joinReqPL.SenderID, nsKEK)

	h.log.WithFields(log.Fields{
		"message_type":   ans.BasePayload.MessageType,
		"sender_id":      ans.BasePayload.SenderID,
		"receiver_id":    ans.BasePayload.ReceiverID,
		"transaction_id": ans.BasePayload.TransactionID,
		"result_code":    ans.Result.ResultCode,
	}).Info("backend/joinserver: sending response")

	h.returnPayload(w, http.StatusOK, ans)
}

func (h *handler) handleRejoinReq(w http.ResponseWriter, b []byte) {
	var rejoinReqPL backend.RejoinReqPayload
	err := json.Unmarshal(b, &rejoinReqPL)
	if err != nil {
		h.returnError(w, http.StatusBadRequest, backend.Other, err.Error())
		return
	}

	dk, err := h.config.GetDeviceKeysByDevEUIFunc(rejoinReqPL.DevEUI)
	if err != nil {
		switch err {
		case ErrDevEUINotFound:
			h.returnRejoinReqError(w, rejoinReqPL.BasePayload, http.StatusBadRequest, backend.UnknownDevEUI, err.Error())
		default:
			h.returnRejoinReqError(w, rejoinReqPL.BasePayload, http.StatusBadRequest, backend.Other, err.Error())
		}
		return
	}

	nsKEK, err := h.config.GetKEKByLabelFunc(rejoinReqPL.SenderID)
	if err != nil {
		h.returnRejoinReqError(w, rejoinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	asKEKLabel, err := h.config.GetASKEKLabelByDevEUIFunc(rejoinReqPL.DevEUI)
	if err != nil {
		h.returnRejoinReqError(w, rejoinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	asKEK, err := h.config.GetKEKByLabelFunc(asKEKLabel)
	if err != nil {
		h.returnRejoinReqError(w, rejoinReqPL.BasePayload, http.StatusInternalServerError, backend.Other, err.Error())
		return
	}

	ans := handleRejoinRequestWrapper(rejoinReqPL, dk, asKEKLabel, asKEK, rejoinReqPL.SenderID, nsKEK)

	h.log.WithFields(log.Fields{
		"message_type":   ans.BasePayload.MessageType,
		"sender_id":      ans.BasePayload.SenderID,
		"receiver_id":    ans.BasePayload.ReceiverID,
		"transaction_id": ans.BasePayload.TransactionID,
		"result_code":    ans.Result.ResultCode,
	}).Info("backend/joinserver: sending response")

	h.returnPayload(w, http.StatusOK, ans)
}
