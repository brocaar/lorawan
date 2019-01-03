package joinserver

import (
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/backend"
)

type context struct {
	joinReqPayload   backend.JoinReqPayload
	joinAnsPayload   backend.JoinAnsPayload
	rejoinReqPayload backend.RejoinReqPayload
	rejoinAnsPaylaod backend.RejoinAnsPayload
	joinType         lorawan.JoinType
	phyPayload       lorawan.PHYPayload
	deviceKeys       DeviceKeys
	devNonce         lorawan.DevNonce
	joinNonce        lorawan.JoinNonce
	netID            lorawan.NetID
	devEUI           lorawan.EUI64
	joinEUI          lorawan.EUI64
	fNwkSIntKey      lorawan.AES128Key
	appSKey          lorawan.AES128Key
	sNwkSIntKey      lorawan.AES128Key
	nwkSEncKey       lorawan.AES128Key
	nsKEKLabel       string
	nsKEK            []byte
	asKEKLabel       string
	asKEK            []byte
}
