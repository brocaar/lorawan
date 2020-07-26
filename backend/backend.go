// Package backend provides the LoRaWAN backend interfaces structs.
package backend

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	keywrap "github.com/NickBall/go-aes-key-wrap"
	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
)

// RatePolicy defines the RatePolicy type.
type RatePolicy string

// RoamingType defines the RoamingType type.
type RoamingType string

// Available rate policies.
const (
	Drop RatePolicy = "Drop"
	Mark RatePolicy = "Mark"
)

// Available roaming activation types.
const (
	Passive  RoamingType = "Passive"
	Handover RoamingType = "Handover"
)

// Supported protocol versions.
const (
	ProtocolVersion1_0 = "1.0"
)

// MessageType defines the message-type type.
type MessageType string

// Supported message types.
const (
	JoinReq     MessageType = "JoinReq"
	JoinAns     MessageType = "JoinAns"
	RejoinReq   MessageType = "RejoinReq"
	RejoinAns   MessageType = "RejoinAns"
	AppSKeyReq  MessageType = "AppSKeyReq"
	AppSKeyAns  MessageType = "AppSKeyAns"
	PRStartReq  MessageType = "PRStartReq"
	PRStartAns  MessageType = "PRStartAns"
	PRStopReq   MessageType = "PRStopReq"
	PRStopAns   MessageType = "PRStopAns"
	HRStartReq  MessageType = "HRStartReq"
	HRStartAns  MessageType = "HRStartAns"
	HRStopReq   MessageType = "HRStopReq"
	HRStopAns   MessageType = "HRStopAns"
	HomeNSReq   MessageType = "HomeNSReq"
	HomeNSAns   MessageType = "HomeNSAns"
	ProfileReq  MessageType = "ProfileReq"
	ProfileAns  MessageType = "ProfileAns"
	XmitDataReq MessageType = "XmitDataReq"
	XmitDataAns MessageType = "XmitDataAns"
)

// ResultCode defines the result-code type.
type ResultCode string

// Supported Result values
const (
	Success                ResultCode = "Success"                // Success, i.e., request was granted
	MICFailed              ResultCode = "MICFailed"              // MIC verification has failed
	JoinReqFailed          ResultCode = "JoinReqFailed"          // JS processing of the JoinReq has failed
	NoRoamingAgreement     ResultCode = "NoRoamingAgreement"     // There is no roaming agreement between the operators
	DevRoamingDisallowed   ResultCode = "DevRoamingDisallowed"   // End-Device is not allowed to roam
	RoamingActDisallowed   ResultCode = "RoamingActDisallowedA"  // End-Device is not allowed to perform activation while roaming
	ActivationDisallowed   ResultCode = "ActivationDisallowed"   // End-Device is not allowed to perform activation
	UnknownDevEUI          ResultCode = "UnknownDevEUI"          // End-Device with a matching DevEUI is not found
	UnknownDevAddr         ResultCode = "UnknownDevAddr"         // End-Device with a matching DevAddr is not found
	UnknownSender          ResultCode = "UnknownSender"          // SenderID is unknown
	UnknownReceiver        ResultCode = "UnkownReceiver"         // ReceiverID is unknown
	Deferred               ResultCode = "Deferred"               // Passive Roaming is not allowed for a period of time
	XmitFailed             ResultCode = "XmitFailed"             // fNS failed to transmit DL packet
	InvalidFPort           ResultCode = "InvalidFPort"           // Invalid FPort for DL (e.g., FPort=0)
	InvalidProtocolVersion ResultCode = "InvalidProtocolVersion" // ProtocolVersion is not supported
	StaleDeviceProfile     ResultCode = "StaleDeviceProfile"     // Device Profile is stale
	MalformedRequest       ResultCode = "MalformedRequest"       // JSON parsing failed (missing object or incorrect content)
	FrameSizeError         ResultCode = "FrameSizeError"         // Wrong size of PHYPayload or FRMPayload
	Other                  ResultCode = "Other"                  // Used for encoding error cases that are not standardized yet
)

// Answer defines the payload answer interface.
type Answer interface {
	// GetBasePayload returns the base payload of the answer.
	GetBasePayload() BasePayloadResult
}

// Request defines the payload request interface.
type Request interface {
	// GetBasePayload returns the base payload of the request.
	GetBasePayload() BasePayload
}

// HEXBytes defines a type which represents bytes as HEX when marshaled to
// text.
type HEXBytes []byte

// String implements fmt.Stringer.
func (hb HEXBytes) String() string {
	return hex.EncodeToString(hb[:])
}

// MarshalText implements encoding.TextMarshaler.
func (hb HEXBytes) MarshalText() ([]byte, error) {
	return []byte(hb.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (hb *HEXBytes) UnmarshalText(text []byte) error {
	b, err := hex.DecodeString(string(text))
	if err != nil {
		return err
	}
	*hb = HEXBytes(b)
	return nil
}

// ISO8601Time defines an ISO 8601 encoded timestamp.
type ISO8601Time time.Time

// MarshalText implements encoding.TextMarshaler.
func (t ISO8601Time) MarshalText() ([]byte, error) {
	return []byte(time.Time(t).Format(time.RFC3339)), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *ISO8601Time) UnmarshalText(text []byte) error {
	ts, err := time.Parse(time.RFC3339, string(text))
	if err != nil {
		return err
	}
	*t = ISO8601Time(ts)
	return nil
}

// Frequency defines the frequency type (in Hz).
type Frequency int

// MarshalJSON implements the json.Marshaler interface.
// This returns the frequency value in MHz (e.g. 868.1) to be compatible
// with the LoRaWAN Backend Interfaces specification.
func (f Frequency) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(f) / 1000000)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This parses a frequency in MHz (float type) back to Hz (int).
func (f *Frequency) UnmarshalJSON(str []byte) error {
	mhz, err := strconv.ParseFloat(string(str), 64)
	if err != nil {
		return errors.Wrap(err, "parse float error")
	}
	*f = Frequency(mhz * 1000000)
	return nil
}

// Percentage defines the percentage type as an int (1 = 1%, 100 = 100%).
type Percentage int

// MarshalJSON implements the json.Marshaler interface.
// This returns the percentage as a float (0.1 for 10%) to be compatible
// with the LoRaWAN Backend Interfaces specification.
func (p Percentage) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(p) / 100)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This parses a percentage presented as 0.1 (float) back to 10 (int).
func (p *Percentage) UnmarshalJSON(str []byte) error {
	perc, err := strconv.ParseFloat(string(str), 64)
	if err != nil {
		return errors.Wrap(err, "parse float error")
	}
	*p = Percentage(perc * 100)
	return nil
}

// BasePayload defines the base payload that is sent with every request.
type BasePayload struct {
	ProtocolVersion string      `json:"ProtocolVersion"` // Version of backend specification. E.g., "1.0"
	SenderID        string      `json:"SenderID"`        // Hexadecimal representation in ASCII format in case of carrying NetID or JoinEUI, ASCII string in case of AS-ID
	ReceiverID      string      `json:"ReceiverID"`      // Hexadecimal representation in ASCII format in case of carrying NetID or JoinEUI, ASCII string in case of AS-ID
	TransactionID   uint32      `json:"TransactionID"`
	MessageType     MessageType `json:"MessageType"`
	SenderToken     HEXBytes    `json:"SenderToken,omitempty"`
	ReceiverToken   HEXBytes    `json:"ReceiverToken,omitempty"`
	VSExtension     VSExtension `json:"VSExtension,omitempty"`
}

// BasePayloadResult defines the base payload that is sent with every result.
type BasePayloadResult struct {
	BasePayload
	Result Result `json:"Result"`
}

// GetBasePayload returns the base payload.
func (p BasePayloadResult) GetBasePayload() BasePayloadResult {
	return p
}

// Result defines the result object.
type Result struct {
	ResultCode  ResultCode `json:"ResultCode"`
	Description string     `json:"Description"` // Detailed information related to the ResultCode (optional).
}

// KeyEnvelope defines the key envelope object.
type KeyEnvelope struct {
	KEKLabel string   `json:"KEKLabel"`
	AESKey   HEXBytes `json:"AESKey"`
}

// Unwrap unwraps the AESKey with the given Key Encryption Key.
func (k KeyEnvelope) Unwrap(kek []byte) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key

	block, err := aes.NewCipher(kek)
	if err != nil {
		return key, errors.Wrap(err, "new cipher error")
	}

	b, err := keywrap.Unwrap(block, k.AESKey[:])
	if err != nil {
		return key, errors.Wrap(err, "unwrap key errror")
	}

	copy(key[:], b)
	return key, nil
}

// NewKeyEnvelope creates a new KeyEnvelope.
func NewKeyEnvelope(kekLabel string, kek []byte, key lorawan.AES128Key) (*KeyEnvelope, error) {
	if kekLabel == "" || len(kek) == 0 {
		return &KeyEnvelope{
			AESKey: HEXBytes(key[:]),
		}, nil
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, errors.Wrap(err, "new cipher error")
	}

	b, err := keywrap.Wrap(block, key[:])
	if err != nil {
		return nil, errors.Wrap(err, "key wrap error")
	}

	return &KeyEnvelope{
		KEKLabel: kekLabel,
		AESKey:   HEXBytes(b),
	}, nil
}

// VSExtension defines vendor specific data.
type VSExtension struct {
	VendorID HEXBytes        `json:"VendorID,omitempty"` // OUI of the vendor
	Object   json.RawMessage `json:"Object,omitempty"`   // The nature of the object is not defined
}

// GWInfoElement defines the gateway info element.
type GWInfoElement struct {
	ID           HEXBytes `json:"ID,omitempty"`           // TODO: shouldn't this be the gateway MAC (64 bit)?
	FineRecvTime *int     `json:"FineRecvTime,omitempty"` // Nanosec within RecvTime
	RFRegion     string   `json:"RFRegion,omitempty"`
	RSSI         *int     `json:"RSSI,omitempty"` // Signed integer, unit: dBm
	SNR          *float64 `json:"SNR,omitempty"`  // Unit: dB
	Lat          *float64 `json:"Lat,omitempty"`
	Lon          *float64 `json:"Lon,omitempty"`
	ULToken      HEXBytes `json:"ULToken,omitempty"`
	DLAllowed    bool     `json:"DLAllowed,omitempty"`
}

// ULMetaData defines the uplink metadata.
type ULMetaData struct {
	DevEUI     *lorawan.EUI64   `json:"DevEUI,omitempty"`
	DevAddr    *lorawan.DevAddr `json:"DevAddr,omitempty"`
	FPort      *uint8           `json:"FPort,omitempty"`
	FCntDown   *uint32          `json:"FCntDown,omitempty"`
	FCntUp     *uint32          `json:"FCntUp,omitempty"`
	Confirmed  bool             `json:"Confirmed,omitempty"`
	DataRate   *int             `json:"DataRate,omitempty"` // See data rate tables in Regional Parameters document
	ULFreq     *float64         `json:"ULFreq,omitempty"`   // Floating point (MHz)
	Margin     *int             `json:"Margin,omitempty"`   // Integer value reported by the end-device in DevStatusAns
	Battery    *int             `json:"Battery,omitempty"`  // Integer value reported by the end-device in DevStatusAns
	FNSULToken HEXBytes         `json:"FNSULToken,omitempty"`
	RecvTime   ISO8601Time      `json:"RecvTime"`
	RFRegion   string           `json:"RFRegion,omitempty"`
	GWCnt      *int             `json:"GWCnt,omitempty"`
	GWInfo     []GWInfoElement  `json:"GWInfo,omitempty"`
}

// DLMetaData defines the downlink metadata.
type DLMetaData struct {
	DevEUI         *lorawan.EUI64  `json:"DevEUI,omitempty"`
	FPort          *uint8          `json:"FPort,omitempty"`
	FCntDown       *uint32         `json:"FCntDown,omitempty"`
	Confirmed      bool            `json:"Confirmed,omitempty"`
	DLFreq1        *float64        `json:"DLFreq1,omitempty"` // TODO: In MHz? At least DLFreq1 or DLFreq2 SHALL be present.
	DLFreq2        *float64        `json:"DLFreq2,omitempty"` // TODO: In Mhz? At least DLFreq1 or DLFreq2 SHALL be present.
	RXDelay1       *int            `json:"RXDelay1,omitempty"`
	ClassMode      *string         `json:"ClassMode,omitempty"` // Only "A" and "C" are supported
	DataRate1      *int            `json:"DataRate1,omitempty"` // Present only if DLFreq1 is present
	DataRate2      *int            `json:"DataRate2,omitempty"` // Present only if DLFreq2 is present
	FNSULToken     HEXBytes        `json:"FNSULToken,omitempty"`
	GWInfo         []GWInfoElement `json:"GWInfo"`
	HiPriorityFlag bool            `json:"HiPriorityFlag,omitempty"`
}

// JoinReqPayload defines the JoinReq message payload.
type JoinReqPayload struct {
	BasePayload
	MACVersion string             `json:"MACVersion"` // e.g. "1.0.2"
	PHYPayload HEXBytes           `json:"PHYPayload"`
	DevEUI     lorawan.EUI64      `json:"DevEUI"`
	DevAddr    lorawan.DevAddr    `json:"DevAddr"`
	DLSettings lorawan.DLSettings `json:"DLSettings"`
	RxDelay    int                `json:"RxDelay"`
	CFList     HEXBytes           `json:"CFList,omitempty"` // Optional
}

// GetBasePayload returns the base payload.
func (p JoinReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// JoinAnsPayload defines the JoinAns message payload.
type JoinAnsPayload struct {
	BasePayloadResult
	PHYPayload   HEXBytes     `json:"PHYPayload,omitempty"`   // Mandatory when Result=Success
	Lifetime     *int         `json:"Lifetime,omitempty"`     // Mandatory when Result=Success, in seconds
	SNwkSIntKey  *KeyEnvelope `json:"SNwkSIntKey,omitempty"`  // Mandatory when Result=Success
	FNwkSIntKey  *KeyEnvelope `json:"FNwkSIntKey,omitempty"`  // Mandatory when Result=Success
	NwkSEncKey   *KeyEnvelope `json:"NwkSEncKey,omitempty"`   // Mandatory when Result=Success
	NwkSKey      *KeyEnvelope `json:"NwkSKey,omitempty"`      // Mandatory when Result=Success (LoRaWAN 1.0.x)
	AppSKey      *KeyEnvelope `json:"AppSKey,omitempty"`      // Mandatory when Result=Success and not SessionKeyID
	SessionKeyID HEXBytes     `json:"SessionKeyID,omitempty"` // Mandatory when Result=Success and not AppSKey
}

// GetBasePayload returns the base payload.
func (p JoinAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// RejoinReqPayload defines the RejoinReq message payload.
type RejoinReqPayload struct {
	BasePayload
	MACVersion string             `json:"MACVersion"` // e.g. "1.0.2"
	PHYPayload HEXBytes           `json:"PHYPayload"`
	DevEUI     lorawan.EUI64      `json:"DevEUI"`
	DevAddr    lorawan.DevAddr    `json:"DevAddr"`
	DLSettings lorawan.DLSettings `json:"DLSettings"`
	RxDelay    int                `json:"RxDelay"`
	CFList     HEXBytes           `json:"CFList,omitempty"` // Optional
}

// GetBasePayload returns the base payload.
func (p RejoinReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// RejoinAnsPayload defines the RejoinAns message payload.
type RejoinAnsPayload struct {
	BasePayloadResult
	PHYPayload   HEXBytes     `json:"PHYPayload,omitempty"`   // Mandatory when Result=Success
	Lifetime     *int         `json:"Lifetime,omitempty"`     // Mandatory when Result=Success, in seconds
	SNwkSIntKey  *KeyEnvelope `json:"SNwkSIntKey,omitempty"`  // Mandatory when Result=Success
	FNwkSIntKey  *KeyEnvelope `json:"FNwkSIntKey,omitempty"`  // Mandatory when Result=Success
	NwkSEncKey   *KeyEnvelope `json:"NwkSEncKey,omitempty"`   // Mandatory when Result=Success
	NwkSKey      *KeyEnvelope `json:"NwkSKey,omitempty"`      // Mandatory when Result=Success (LoRaWAN 1.0.x)
	AppSKey      *KeyEnvelope `json:"AppSKey,omitempty"`      // Mandatory when Result=Success and not SessionKeyID
	SessionKeyID HEXBytes     `json:"SessionKeyID,omitempty"` // Mandatory when Result=Success and not AppSKey
}

// GetBasePayload returns the base payload.
func (p RejoinAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// AppSKeyReqPayload defines the AppSKeyReq message payload.
type AppSKeyReqPayload struct {
	BasePayload
	DevEUI       lorawan.EUI64 `json:"DevEUI"`
	SessionKeyID HEXBytes      `json:"SessionKeyID"`
}

// GetBasePayload returns the base payload.
func (p AppSKeyReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// AppSKeyAnsPayload defines the AppSKeyAns message payload.
type AppSKeyAnsPayload struct {
	BasePayloadResult
	DevEUI       lorawan.EUI64 `json:"DevEUI"`
	AppSKey      *KeyEnvelope  `json:"AppSKey,omitempty"` // Mandatory when Result=Success
	SessionKeyID HEXBytes      `json:"SessionKeyID"`
}

// GetBasePayload returns the base payload.
func (p AppSKeyAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// PRStartReqPayload defines the PRStartReq message payload.
type PRStartReqPayload struct {
	BasePayload
	PHYPayload HEXBytes   `json:"PHYPayload,omitempty"`
	ULMetaData ULMetaData `json:"ULMetaData"`
}

// GetBasePayload returns the base payload.
func (p PRStartReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// PRStartAnsPayload defines the PRStartAns message payload.
type PRStartAnsPayload struct {
	BasePayloadResult
	PHYPayload     HEXBytes         `json:"PHYPayload,omitempty"`     // Optional when Result=Success
	DevEUI         *lorawan.EUI64   `json:"DevEUI,omitempty"`         // Optional when Result=Success
	Lifetime       *int             `json:"Lifetime,omitempty"`       // Mandatory when Result=Success, in seconds
	FNwkSIntKey    *KeyEnvelope     `json:"FNwkSIntKey,omitempty"`    // Optional when Result=Success and not NwkSKey
	NwkSKey        *KeyEnvelope     `json:"NwkSKey,omitempty"`        // Optional when Result=Success and not FNwkSIntKey
	FCntUp         *uint32          `json:"FCntUp,omitempty"`         // Optional when Result=Success
	ServiceProfile *ServiceProfile  `json:"ServiceProfile,omitempty"` // Optional when Result=Success
	DLMetaData     *DLMetaData      `json:"DLMetaData,omitempty"`     // Optional when Result=Success
	DevAddr        *lorawan.DevAddr `json:"DevAddr,omitempty"`        // Optional when Result=Success (not specified in Backend Specs but needed for OTAA)
}

// GetBasePayload returns the base payload.
func (p PRStartAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// PRStopReqPayload defines the PRStopReq message payload.
type PRStopReqPayload struct {
	BasePayload
	DevEUI   lorawan.EUI64 `json:"DevEUI"`
	Lifetime *int          `json:"Lifetime,omitempty"` // Optional, in seconds
}

// GetBasePayload returns the base payload.
func (p PRStopReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// PRStopAnsPayload defines the PRStopAns message payload.
type PRStopAnsPayload struct {
	BasePayloadResult
}

// GetBasePayload returns the base payload.
func (p PRStopAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// HRStartReqPayload defines the HRStartReq message payload.
type HRStartReqPayload struct {
	BasePayload
	MACVersion             string             `json:"MACVersion"` // e.g. "1.0.2"
	PHYPayload             HEXBytes           `json:"PHYPayload"`
	DevAddr                lorawan.DevAddr    `json:"DevAddr"`
	DeviceProfile          DeviceProfile      `json:"DeviceProfile"`
	ULMetaData             ULMetaData         `json:"ULMetaData"`
	DLSettings             lorawan.DLSettings `json:"DLSettings"`
	RxDelay                int                `json:"RxDelay"`
	CFList                 HEXBytes           `json:"CFList,omitempty"`       // Optional
	DeviceProfileTimestamp ISO8601Time        `json:"DeviceProfileTimestamp"` // Timestamp of last DeviceProfile change
}

// GetBasePayload returns the base payload.
func (p HRStartReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// HRStartAnsPayload defines the HRStartAns message payload.
type HRStartAnsPayload struct {
	BasePayloadResult
	PHYPayload             HEXBytes        `json:"PHYPayload,omitempty"`             // Mandatory when Result=Success
	Lifetime               *int            `json:"Lifetime,omitempty"`               // Mandatory when Result=Success, in seconds
	SNwkSIntKey            *KeyEnvelope    `json:"SNwkSIntKey,omitempty"`            // Mandatory when Result=Success
	FNwkSIntKey            *KeyEnvelope    `json:"FNwkSIntKey,omitempty"`            // Mandatory when Result=Success
	NwkSEncKey             *KeyEnvelope    `json:"NwkSEncKey,omitempty"`             // Mandatory when Result=Success
	NwkSKey                *KeyEnvelope    `json:"NwkSKey,omitempty"`                // Mandatory when Result=Success (LoRaWAN 1.0.x)
	DeviceProfile          *DeviceProfile  `json:"DeviceProfile,omitempty"`          // Optional, when Result=Failure
	ServiceProfile         *ServiceProfile `json:"ServiceProfile,omitempty"`         // Mandatory when Result=Success
	DLMetaData             *DLMetaData     `json:"DLMetaData,omitempty"`             // Mandatory when Result=Success
	DeviceProfileTimestamp *ISO8601Time    `json:"DeviceProfileTimestamp,omitempty"` // Optional, when Result=Failure, timestamp of last DeviceProfile change
}

// GetBasePayload returns the base payload.
func (p HRStartAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// HRStopReqPayload defines the HRStopReq message payload.
type HRStopReqPayload struct {
	BasePayload
	DevEUI lorawan.EUI64 `json:"DevEUI"`
}

// GetBasePayload returns the base payload.
func (p HRStopReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// HRStopAnsPayload defines the HRStopAns message payload.
type HRStopAnsPayload struct {
	BasePayloadResult
}

// GetBasePayload returns the base payload.
func (p HRStopAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// HomeNSReqPayload defines the HomeNSReq message payload.
type HomeNSReqPayload struct {
	BasePayload
	DevEUI lorawan.EUI64 `json:"DevEUI"`
}

// GetBasePayload returns the base payload.
func (p HomeNSReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// HomeNSAnsPayload defines the HomeNSAns message payload.
type HomeNSAnsPayload struct {
	BasePayloadResult
	HNetID lorawan.NetID `json:"HNetID"`
}

// GetBasePayload returns the base payload.
func (p HomeNSAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// ProfileReqPayload defines the ProfileReq message payload.
type ProfileReqPayload struct {
	BasePayload
	DevEUI lorawan.EUI64 `json:"DevEUI"`
}

// GetBasePayload returns the base payload.
func (p ProfileReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// ProfileAnsPayload defines the ProfileAns message payload.
type ProfileAnsPayload struct {
	BasePayloadResult
	DeviceProfile          *DeviceProfile `json:"DeviceProfile,omitempty"`          // Mandatory when Result=Success
	DeviceProfileTimestamp *ISO8601Time   `json:"DeviceProfileTimestamp,omitempty"` // Mandatory when Result=Success. Timestamp of last DeviceProfile change.
	RoamingActivationType  *RoamingType   `json:"RoamingActivationType"`            // Mandatory when Result=Success.
}

// GetBasePayload returns the base payload.
func (p ProfileAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// XmitDataReqPayload defines the XmitDataReq message payload.
type XmitDataReqPayload struct {
	BasePayload
	PHYPayload HEXBytes    `json:"PHYPayload,omitempty"` // Either PHYPayload or FRMPayload should be used
	FRMPayload HEXBytes    `json:"FRMPayload,omitempty"` // Either PHYPayload or FRMPayload should be used
	ULMetaData *ULMetaData `json:"ULMetaData,omitempty"` // Either ULMetaData or DLMetaData must be used
	DLMetaData *DLMetaData `json:"DLMetaData,omitempty"` // Either ULMetaData or DLMetaData must be used
}

// GetBasePayload returns the base payload.
func (p XmitDataReqPayload) GetBasePayload() BasePayload {
	return p.BasePayload
}

// XmitDataAnsPayload defines the XmitDataAns message payload.
type XmitDataAnsPayload struct {
	BasePayloadResult
	DLFreq1 *float64 `json:"DLFreq1,omitempty"` // Optional, when Result=Success, TODO: In MHz?
	DLFreq2 *float64 `json:"DLFreq2,omitempty"` // Optional, when Result=Success, TODO: In Mhz?
}

// GetBasePayload returns the base payload.
func (p XmitDataAnsPayload) GetBasePayload() BasePayloadResult {
	return p.BasePayloadResult
}

// ServiceProfile includes service parameters that are needed by the NS for
// setting up the LoRa radio access service and interfacing with the AS.
type ServiceProfile struct {
	ServiceProfileID       string     `json:"ServiceProfile" db:"service_profile_id"`
	ULRate                 int        `json:"ULRate" db:"ul_rate"`
	ULBucketSize           int        `json:"ULBucketSize" db:"ul_bucket_size"`
	ULRatePolicy           RatePolicy `json:"ULRatePolicy" db:"ul_rate_policy"`
	DLRate                 int        `json:"DLRate" db:"dl_rate"`
	DLBucketSize           int        `json:"DLBucketSize" db:"dl_bucket_size"`
	DLRatePolicy           RatePolicy `json:"DLRatePolicy" db:"dl_rate_policy"`
	AddGWMetadata          bool       `json:"AddGWMetadata" db:"add_gw_metadata"`
	DevStatusReqFreq       int        `json:"DevStatusReqFreq" db:"dev_status_req_freq"`            // Unit: requests-per-day
	ReportDevStatusBattery bool       `json:"ReportDevStatusBatery" db:"report_dev_status_battery"` // TODO: there is a typo in the spec!
	ReportDevStatusMargin  bool       `json:"ReportDevStatusMargin" db:"report_dev_status_margin"`
	DRMin                  int        `json:"DRMin" db:"dr_min"`
	DRMax                  int        `json:"DRMax" db:"dr_max"`
	ChannelMask            HEXBytes   `json:"ChannelMask" db:"channel_mask"`
	PRAllowed              bool       `json:"PRAllowed" db:"pr_allowed"`
	HRAllowed              bool       `json:"HRAllowed" db:"hr_allowed"`
	RAAllowed              bool       `json:"RAAAllowed" db:"ra_allowed"`
	NwkGeoLoc              bool       `json:"NwkGeoLoc" db:"nwk_geo_loc"`
	TargetPER              Percentage `json:"TargetPER" db:"target_per"` // Example: 0.10 indicates 10%
	MinGWDiversity         int        `json:"MinGWDiversity" db:"min_gw_diversity"`
}

// DeviceProfile includes End-Device capabilities and boot parameters that are
// needed by the NS for setting up the LoRaWAN radio access service. These
// information elements SHALL be provided by the End-Device manufacturer.
type DeviceProfile struct {
	DeviceProfileID    string      `json:"DeviceProfileID" db:"device_profile_id"`
	SupportsClassB     bool        `json:"SupportsClassB" db:"supports_class_b"`
	ClassBTimeout      int         `json:"ClassBTimeout" db:"class_b_timeout"` // Unit: seconds
	PingSlotPeriod     int         `json:"PingSlotPeriod" db:"ping_slot_period"`
	PingSlotDR         int         `json:"PingSLotDR" db:"ping_slot_dr"`
	PingSlotFreq       Frequency   `json:"PingSlotFreq" db:"ping_slot_freq"` // TODO: in MHz?
	SupportsClassC     bool        `json:"SupportsClassC" db:"supports_class_c"`
	ClassCTimeout      int         `json:"ClassCTimeout" db:"class_c_timeout"`         // Unit: seconds
	MACVersion         string      `json:"MACVersion" db:"mac_version"`                // Example: "1.0.2" [LW102]
	RegParamsRevision  string      `json:"RegParamsRevision" db:"reg_params_revision"` // Example: "B" [RP102B]
	RXDelay1           int         `json:"RXDelay1" db:"rx_delay_1"`
	RXDROffset1        int         `json:"RXDROffset1" db:"rx_dr_offset_1"`
	RXDataRate2        int         `json:"RXDataRate2" db:"rx_data_rate_2"`              // Unit: bits-per-second
	RXFreq2            Frequency   `json:"RXFreq2" db:"rx_freq_2"`                       // Value of the frequency, e.g., 868.10
	FactoryPresetFreqs []Frequency `json:"FactoryPresetFreqs" db:"factory_preset_freqs"` // TODO: In MHz?
	MaxEIRP            int         `json:"MaxEIRP" db:"max_eirp"`                        // In dBm
	MaxDutyCycle       Percentage  `json:"MaxDutyCycle" db:"max_duty_cycle"`             // Example: 0.10 indicates 10%
	SupportsJoin       bool        `json:"SupportsJoin" db:"supports_join"`
	RFRegion           string      `json:"RFRegion" db:"rf_region"`
	Supports32bitFCnt  bool        `json:"Supports32bitFCnt" db:"supports_32bit_fcnt"`
}

// RoutingProfile includes information that are needed by the NS for setting
// up data-plane with the AS.
type RoutingProfile struct {
	RoutingProfileID string `json:"RoutingProfileID" db:"routing_profile_id"`
	ASID             string `json:"AS-ID" db:"as_id"` // Value can be IP address, DNS name, etc.
}

// NetworkActivationRecord is used for keeping track of the End-Devices
// performing Activation away from Home. When the Activation away from Home
// Procedure takes place, then the NS SHALL generate a monthly Network
// Activation Record for each ServiceProfileID of another NS that has at least
// one End-Device active throughout the month, and dedicated Network Activation
// Records for each activation and deactivation of an End-Device from another
// NS.
type NetworkActivationRecord struct {
	NetID              lorawan.NetID `db:"net_id"`               // NetID of the roaming partner NS
	ServiceProfileID   string        `db:"service_profile_id"`   // Service Profile ID
	IndividualRecord   bool          `db:"individual_record"`    // Indicates if this is an individual (de-)activation record (as opposed to cumulative record of End-Devices that are active throughout the month)
	TotalActiveDevices int           `db:"total_active_devices"` // Number of End-Devices that have been active throughout the month. Included if this is a cumulative record.
	DevEUI             lorawan.EUI64 `db:"dev_eui"`              // DevEUI of the End-Device that has performed the (de-)activation. Included if this is an IndividualRecord for a (de-)activation event.
	ActivationTime     time.Time     `db:"activation_time"`      // Date/time of the activation. Included if this is an IndividualRecord for an activation event.
	DeactivationTime   time.Time     `db:"deactivation_time"`    // Date/time of the deactivation. Included if this is an IndividualRecord for a deactivation event.
}

// NetworkTrafficRecord is used for keeping track of the amount of traffic
// served for roaming End-Devices. The NS that allows roaming SHALL generate
// a monthly Network Traffic Record for each roaming type (Passive/Handover
// Roaming) under each ServiceProfileID of another NS that has at least one
// End-Device roaming into its network.
//
// Packet and payload counters are only based on the user-generated traffic.
// Payload counters are based on the size of the FRMPayload field.
type NetworkTrafficRecord struct {
	NetID                    lorawan.NetID `db:"net_id"`                       // NetID of the roaming partner NS
	ServiceProfileID         string        `db:"service_profile_id"`           // Service Profile ID
	RoamingType              RoamingType   `db:"roaming_type"`                 // Passive Roaming or Handover Roaming
	TotalULPackets           int           `db:"total_ul_packets"`             // Number of uplink packets
	TotalDLPackets           int           `db:"total_dl_packets"`             // Number of downlink packets
	TotalOutProfileULPackets int           `db:"total_out_profile_ul_packets"` // Number of uplink packets that exceeded ULRate but forwarded anyways per ULRatePolicy
	TotalOutProfileDLPackets int           `db:"total_out_profile_dl_packets"` // Number of downlink packets that exceeded DLRate but forwarded anyways per DLRatePolicy
	TotalULBytes             int           `db:"total_ul_bytes"`               // Total amount of uplink bytes
	TotalDLBytes             int           `db:"total_dl_bytes"`               // Total amount of downlink bytes
	TotalOutProfileULBytes   int           `db:"total_out_profile_ul_bytes"`   // Total amount of uplink bytes that falls outside the Service Profile
	TotalOutProfileDLBytes   int           `db:"total_out_profile_dl_bytes"`   // Total amount of downlink bytes that falls outside the Service Profile
}
