package joinserver

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

var rejoinTasks = []func(*context) error{
	setRejoinContext,
	setJoinNonce,
	setSessionKeys,
	createRejoinAnsPayload,
}

func handleRejoinRequestWrapper(rejoinReqPL backend.RejoinReqPayload, dk DeviceKeys, asKEKLabel string, asKEK []byte, nsKEKLabel string, nsKEK []byte) backend.RejoinAnsPayload {
	basePayload := backend.BasePayload{
		ProtocolVersion: backend.ProtocolVersion1_0,
		SenderID:        rejoinReqPL.ReceiverID,
		ReceiverID:      rejoinReqPL.SenderID,
		TransactionID:   rejoinReqPL.TransactionID,
		MessageType:     backend.RejoinAns,
	}

	rjaPL, err := handleRejoinRequest(rejoinReqPL, dk, asKEKLabel, asKEK, nsKEKLabel, nsKEK)
	if err != nil {
		var resCode backend.ResultCode

		switch errors.Cause(err) {
		case ErrInvalidMIC:
			resCode = backend.MICFailed
		default:
			resCode = backend.Other
		}

		rjaPL = backend.RejoinAnsPayload{
			BasePayload: basePayload,
			Result: backend.Result{
				ResultCode:  resCode,
				Description: err.Error(),
			},
		}
	}

	rjaPL.BasePayload = basePayload
	return rjaPL
}

func handleRejoinRequest(rejoinReqPL backend.RejoinReqPayload, dk DeviceKeys, asKEKLabel string, asKEK []byte, nsKEKLabel string, nsKEK []byte) (backend.RejoinAnsPayload, error) {
	ctx := context{
		rejoinReqPayload: rejoinReqPL,
		deviceKeys:       dk,
		asKEKLabel:       asKEKLabel,
		asKEK:            asKEK,
		nsKEKLabel:       nsKEKLabel,
		nsKEK:            nsKEK,
	}

	for _, f := range rejoinTasks {
		if err := f(&ctx); err != nil {
			return ctx.rejoinAnsPaylaod, err
		}
	}

	return ctx.rejoinAnsPaylaod, nil
}

func setRejoinContext(ctx *context) error {
	if err := ctx.phyPayload.UnmarshalBinary(ctx.rejoinReqPayload.PHYPayload[:]); err != nil {
		return errors.Wrap(err, "unmarshal phypayload error")
	}

	if err := ctx.netID.UnmarshalText([]byte(ctx.rejoinReqPayload.SenderID)); err != nil {
		return errors.Wrap(err, "unmarshal netid error")
	}

	if err := ctx.joinEUI.UnmarshalText([]byte(ctx.rejoinReqPayload.ReceiverID)); err != nil {
		return errors.Wrap(err, "unmarshal joineui error")
	}

	switch v := ctx.phyPayload.MACPayload.(type) {
	case *lorawan.RejoinRequestType02Payload:
		ctx.joinType = v.RejoinType
		ctx.devNonce = lorawan.DevNonce(v.RJCount0)
	case *lorawan.RejoinRequestType1Payload:
		ctx.joinType = v.RejoinType
		ctx.devNonce = lorawan.DevNonce(v.RJCount1)
	default:
		return fmt.Errorf("expected rejoin payload, got %T", ctx.phyPayload.MACPayload)
	}

	ctx.devEUI = ctx.rejoinReqPayload.DevEUI

	return nil
}

func createRejoinAnsPayload(ctx *context) error {
	var cFList *lorawan.CFList
	if len(ctx.rejoinReqPayload.CFList[:]) != 0 {
		cFList = new(lorawan.CFList)
		if err := cFList.UnmarshalBinary(ctx.rejoinReqPayload.CFList[:]); err != nil {
			return errors.Wrap(err, "unmarshal cflist error")
		}
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.JoinAccept,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.JoinAcceptPayload{
			JoinNonce:  ctx.joinNonce,
			HomeNetID:  ctx.netID,
			DevAddr:    ctx.rejoinReqPayload.DevAddr,
			DLSettings: ctx.rejoinReqPayload.DLSettings,
			RXDelay:    uint8(ctx.rejoinReqPayload.RxDelay),
			CFList:     cFList,
		},
	}

	jsIntKey, err := getJSIntKey(ctx.deviceKeys.NwkKey, ctx.devEUI)
	if err != nil {
		return err
	}

	jsEncKey, err := getJSEncKey(ctx.deviceKeys.NwkKey, ctx.devEUI)
	if err != nil {
		return err
	}

	if err := phy.SetDownlinkJoinMIC(ctx.joinType, ctx.joinEUI, ctx.devNonce, jsIntKey); err != nil {
		return err
	}

	if err := phy.EncryptJoinAcceptPayload(jsEncKey); err != nil {
		return err
	}

	b, err := phy.MarshalBinary()
	if err != nil {
		return err
	}

	// as the rejoin-request is only implemented for LoRaWAN1.1+ there is no
	// need to check the OptNeg flag
	ctx.rejoinAnsPaylaod = backend.RejoinAnsPayload{
		Result: backend.Result{
			ResultCode: backend.Success,
		},
		PHYPayload: backend.HEXBytes(b),
		// TODO: add Lifetime?
	}

	ctx.rejoinAnsPaylaod.AppSKey, err = getKeyEnvelope(ctx.asKEKLabel, ctx.asKEK, ctx.appSKey)
	if err != nil {
		return err
	}

	ctx.rejoinAnsPaylaod.FNwkSIntKey, err = getKeyEnvelope(ctx.nsKEKLabel, ctx.nsKEK, ctx.fNwkSIntKey)
	if err != nil {
		return err
	}
	ctx.rejoinAnsPaylaod.SNwkSIntKey, err = getKeyEnvelope(ctx.nsKEKLabel, ctx.nsKEK, ctx.sNwkSIntKey)
	if err != nil {
		return err
	}
	ctx.rejoinAnsPaylaod.NwkSEncKey, err = getKeyEnvelope(ctx.nsKEKLabel, ctx.nsKEK, ctx.nwkSEncKey)
	if err != nil {
		return err
	}

	return nil
}
