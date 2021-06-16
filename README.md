# LoRaWAN (Go)

![Tests](https://github.com/brocaar/lorawan/actions/workflows/main.yml/badge.svg?branch=master)
[![GoDoc](https://godoc.org/github.com/brocaar/lorawan?status.svg)](https://godoc.org/github.com/brocaar/lorawan)

Package lorawan provides structures and tools to read and write LoRaWAN
1.0 and 1.1 frames from and to a slice of bytes.

The following structures are implemented (+ fields):

```
PHYPayload    (MHDR | MACPayload | MIC)
MACPayload    (FHDR | FPort | FRMPayload)
FHDR          (DevAddr | FCtrl | FCnt | FOpts)
```

The Following message types (MType) are implemented:

* JoinRequest
* RejoinRequest
* JoinAccept
* UnconfirmedDataUp
* UnconfirmedDataDown
* ConfirmedDataUp
* ConfirmedDataDown
* Proprietary

The following MAC commands (and their optional payloads) are implemented:

* ResetInd
* ResetConf
* LinkCheckReq
* LinkCheckAns
* LinkADRReq
* LinkADRAns
* DutyCycleReq
* DutyCycleAns
* RXParamSetupReq
* RXParamSetupAns
* DevStatusReq
* DevStatusAns
* NewChannelReq
* NewChannelAns
* RXTimingSetupReq
* RXTimingSetupAns
* TXParamSetupReq
* TXParamSetupAns
* DLChannelReq
* DLChannelAns
* RekeyInd
* RekeyConf
* ADRParamSetupReq
* ADRParamSetupAns
* DeviceTimeReq
* DeviceTimeAns
* ForceRejoinReq
* RejoinParamSetupReq
* RejoinParamSetupAns
* PingSlotInfoReq
* PingSlotInfoAns
* PingSlotChannelReq
* PingSlotChannelAns
* BeaconFreqReq
* BeaconFreqAns
* DeviceModeInd
* DeviceModeConf
* Proprietary commands (0x80 - 0xFF) can be registered with RegisterProprietaryMACCommand


## Sub-packages

* `airtime` functions for calculating TX time-on-air
* `band` ISM band configuration from the LoRaWAN Regional Parameters specification
* `backend` Structs matching the LoRaWAN Backend Interface specification object
* `backend/joinserver` LoRaWAN Backend Interface join-server interface implementation (`http.Handler`)
* `applayer/clocksync` Application Layer Clock Synchronization over LoRaWAN
* `applayer/multicastsetup` Application Layer Remote Multicast Setup over LoRaWAN
* `applayer/fragmentation` Fragmented Data Block Transport over LoRaWAN
* `gps` functions to handle Time <> GPS Epoch time conversion

## Documentation

See https://godoc.org/github.com/brocaar/lorawan. There is also an [examples](https://godoc.org/github.com/brocaar/lorawan#pkg-examples)
section with usage examples. When using this package, knowledge about the LoRaWAN specification is needed.
You can download the LoRaWAN specification here: https://lora-alliance.org/lorawan-for-developers

## Support

For questions, feedback or support, please refer to the ChirpStack Community Forum:
[https://forum.chirpstack.io](https://forum.chirpstack.io/).

## License

This package is distributed under the MIT license which can be found in ``LICENSE``.
LoRaWAN is a trademark of the LoRa Alliance Inc. (https://www.lora-alliance.org/).
