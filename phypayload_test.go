package lorawan

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAES128Key(t *testing.T) {
	Convey("Given an empty AES128Key", t, func() {
		var key AES128Key

		Convey("When the value is [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}", func() {
			key = [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}

			Convey("Then MarshalText returns 01020304050607080102030405060708", func() {
				b, err := key.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "01020304050607080102030405060708")
			})

			Convey("Then Value returns the expected value", func() {
				v, err := key.Value()
				So(err, ShouldBeNil)
				So(v, ShouldResemble, driver.Value(key[:]))
			})
		})

		Convey("Given the string 01020304050607080102030405060708", func() {
			str := "01020304050607080102030405060708"
			Convey("Then UnmarshalText returns AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}", func() {
				err := key.UnmarshalText([]byte(str))
				So(err, ShouldBeNil)
				So(key, ShouldResemble, AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})
			})
		})

		Convey("Given []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}", func() {
			b := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
			Convey("Then Scan scans the value correctly.", func() {
				So(key.Scan(b), ShouldBeNil)
				So(key[:], ShouldResemble, b)
			})
		})
	})
}

func TestMHDR(t *testing.T) {
	Convey("Given an empty MHDR", t, func() {
		var h MHDR
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := h.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given MType=Proprietary, Major=LoRaWANR1", func() {
			h.MType = Proprietary
			h.Major = LoRaWANR1
			Convey("Then MarshalBinary returns []byte{224}", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{224})
			})
		})

		Convey("Given a slice []byte{224}", func() {
			b := []byte{224}
			Convey("Then UnmarshalBinary returns a MHDR with MType=Proprietary, Major=LoRaWANR1", func() {
				err := h.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(h, ShouldResemble, MHDR{MType: Proprietary, Major: LoRaWANR1})
			})
		})
	})
}

func TestPHYPayloadMACPayloadLoRaWAN10(t *testing.T) {
	Convey("Given a set of test for LoRaWAN 1.0", t, func() {
		var fPort1 uint8 = 1
		var fPort0 uint8

		mac1 := MACCommand{
			CID:     LinkCheckReq,
			Payload: nil,
		}
		mac2 := MACCommand{
			CID: LinkADRAns,
			Payload: &LinkADRAnsPayload{
				ChannelMaskACK: true,
				DataRateACK:    false,
				PowerACK:       true,
			},
		}

		testTable := []struct {
			Name       string
			PHYPayload PHYPayload
			NwkSEncKey AES128Key
			AppSKey    AES128Key
			Bytes      []byte
		}{
			{
				Name: "FRMPayload data",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FCtrl: FCtrl{
								ADR: true,
							},
							FCnt: 1,
						},
						FPort: &fPort1,
						FRMPayload: []Payload{
							&DataPayload{Bytes: []byte("hello")},
						},
					},
					MIC: MIC{214, 195, 181, 130},
				},
				NwkSEncKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				AppSKey:    AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:      []byte{64, 4, 3, 2, 1, 128, 1, 0, 1, 166, 148, 100, 38, 21, 214, 195, 181, 130},
			},
			{
				Name: "Mac-commands in FOpts",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FOpts: []Payload{
								&mac1,
								&mac2,
							},
						},
						FPort: &fPort1,
						FRMPayload: []Payload{
							&DataPayload{Bytes: []byte{1, 2, 3, 4}},
						},
					},
					MIC: MIC{182, 77, 192, 57},
				},
				NwkSEncKey: AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				AppSKey:    AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:      []byte{64, 4, 3, 2, 1, 3, 0, 0, 2, 3, 5, 1, 106, 55, 152, 245, 182, 77, 192, 57},
			},
			{
				Name: "Mac-commands in FRMPayload",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FPort: &fPort0,
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
						},
						FRMPayload: []Payload{
							&mac1,
							&mac2,
						},
					},
					MIC: MIC{238, 106, 165, 8},
				},
				NwkSEncKey: AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				AppSKey:    AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:      []byte{64, 4, 3, 2, 1, 0, 0, 0, 0, 105, 54, 158, 238, 106, 165, 8},
			},
		}

		for i, test := range testTable {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				var phy PHYPayload
				So(phy.UnmarshalBinary(test.Bytes), ShouldBeNil)

				var ok bool
				var err error

				switch phy.MHDR.MType {
				case UnconfirmedDataUp, ConfirmedDataUp:
					ok, err = phy.ValidateUplinkDataMIC(LoRaWAN1_0, 0, 0, 0, test.NwkSEncKey, AES128Key{})
				case UnconfirmedDataDown, ConfirmedDataDown:
					ok, err = phy.ValidateDownlinkDataMIC(LoRaWAN1_0, 0, test.NwkSEncKey)
				default:
					t.Fatalf("unexpected MType %s", phy.MHDR.MType)
				}
				So(err, ShouldBeNil)
				So(ok, ShouldBeTrue)

				So(phy.DecodeFOptsToMACCommands(), ShouldBeNil)
				So(phy.DecryptFRMPayload(test.AppSKey), ShouldBeNil)
				if macPL, ok := phy.MACPayload.(*MACPayload); ok {
					macPL.FHDR.FCtrl.fOptsLen = 0
				}
				So(phy, ShouldResemble, test.PHYPayload)

				So(test.PHYPayload.EncryptFRMPayload(test.AppSKey), ShouldBeNil)

				switch test.PHYPayload.MHDR.MType {
				case UnconfirmedDataUp, ConfirmedDataUp:
					err = test.PHYPayload.SetUplinkDataMIC(LoRaWAN1_0, 0, 0, 0, test.NwkSEncKey, AES128Key{})
				case UnconfirmedDataDown, ConfirmedDataDown:
					err = test.PHYPayload.SetDownlinkDataMIC(LoRaWAN1_0, 0, test.NwkSEncKey)
				default:
					t.Fatalf("unexpected MType %s", test.PHYPayload.MHDR.MType)
				}
				So(err, ShouldBeNil)

				b, err := test.PHYPayload.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.Bytes)
			})
		}
	})
}

func TestPHYPayloadMACPayloadLoRaWAN11(t *testing.T) {
	Convey("Given a set of tests for LoRaWAN 1.1", t, func() {
		var fPort1 uint8 = 1
		var fPort0 uint8

		mac1 := MACCommand{
			CID:     LinkCheckReq,
			Payload: nil,
		}
		mac2 := MACCommand{
			CID: LinkADRAns,
			Payload: &LinkADRAnsPayload{
				ChannelMaskACK: true,
				DataRateACK:    false,
				PowerACK:       true,
			},
		}

		testTable := []struct {
			Name         string
			PHYPayload   PHYPayload
			SNwkSIntKey  AES128Key
			FNwkSIntKey  AES128Key
			NwkSEncKey   AES128Key
			AppSKey      AES128Key
			Bytes        []byte
			EncryptFOpts bool
		}{
			{
				Name: "FRMPayload data",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FCtrl: FCtrl{
								ADR: true,
							},
							FCnt: 1,
						},
						FPort: &fPort1,
						FRMPayload: []Payload{
							&DataPayload{Bytes: []byte("hello")},
						},
					},
					MIC: MIC{118, 18, 54, 106},
				},
				SNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				FNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
				NwkSEncKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4},
				AppSKey:     AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:       []byte{64, 4, 3, 2, 1, 128, 1, 0, 1, 166, 148, 100, 38, 21, 118, 18, 54, 106},
			},
			{
				Name: "FRMPayload data with ACK (in this case the confirmed fCnt is used in the mic)",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FCtrl: FCtrl{
								ADR: true,
								ACK: true,
							},
							FCnt: 1,
						},
						FPort: &fPort1,
						FRMPayload: []Payload{
							&DataPayload{Bytes: []byte("hello")},
						},
					},
					MIC: MIC{248, 66, 196, 185},
				},
				SNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				FNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
				NwkSEncKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4},
				AppSKey:     AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:       []byte{64, 4, 3, 2, 1, 160, 1, 0, 1, 166, 148, 100, 38, 21, 248, 66, 196, 185},
			},
			{
				Name: "Mac-commands in FOpts (encrypted, using NFCntDown)",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataDown,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FOpts: []Payload{
								&MACCommand{
									CID: LinkCheckAns,
									Payload: &LinkCheckAnsPayload{
										Margin: 7,
										GwCnt:  1,
									},
								},
							},
						},
					},
					MIC: MIC{226, 79, 31, 159},
				},
				SNwkSIntKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				FNwkSIntKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
				NwkSEncKey:   AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4},
				Bytes:        []byte{96, 4, 3, 2, 1, 3, 0, 0, 223, 180, 241, 226, 79, 31, 159},
				EncryptFOpts: true,
			},
			{
				Name: "Mac-commands in FOpts (encrypted, using AFCntDown encryption flag)",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataDown,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
							FOpts: []Payload{
								&MACCommand{
									CID: LinkCheckAns,
									Payload: &LinkCheckAnsPayload{
										Margin: 7,
										GwCnt:  1,
									},
								},
							},
						},
						FPort: &fPort1,
					},
					MIC: MIC{119, 112, 30, 163},
				},
				SNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				FNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
				NwkSEncKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4},
				Bytes:       []byte{96, 4, 3, 2, 1, 3, 0, 0, 2, 7, 1, 1, 119, 112, 30, 163},
			},
			{
				Name: "Mac-commands in FRMPayload",
				PHYPayload: PHYPayload{
					MHDR: MHDR{
						MType: UnconfirmedDataUp,
						Major: LoRaWANR1,
					},
					MACPayload: &MACPayload{
						FPort: &fPort0,
						FHDR: FHDR{
							DevAddr: DevAddr{1, 2, 3, 4},
						},
						FRMPayload: []Payload{
							&mac1,
							&mac2,
						},
					},
					MIC: MIC{250, 147, 27, 215},
				},
				SNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				FNwkSIntKey: AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 3},
				NwkSEncKey:  AES128Key{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 4},
				AppSKey:     AES128Key{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				Bytes:       []byte{64, 4, 3, 2, 1, 0, 0, 0, 0, 105, 54, 158, 250, 147, 27, 215},
			},
		}

		for i, test := range testTable {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				var phy PHYPayload
				So(phy.UnmarshalBinary(test.Bytes), ShouldBeNil)

				var ok bool
				var err error

				switch phy.MHDR.MType {
				case UnconfirmedDataUp, ConfirmedDataUp:
					ok, err = phy.ValidateUplinkDataMIC(LoRaWAN1_1, 1, 2, 3, test.FNwkSIntKey, test.SNwkSIntKey)
				case UnconfirmedDataDown, ConfirmedDataDown:
					ok, err = phy.ValidateDownlinkDataMIC(LoRaWAN1_1, 1, test.SNwkSIntKey)
				default:
					t.Fatalf("unexpected MType %s", phy.MHDR.MType)
				}
				So(err, ShouldBeNil)
				if !ok {
					var mic MIC
					switch phy.MHDR.MType {
					case UnconfirmedDataUp, ConfirmedDataUp:
						mic, err = phy.calculateUplinkDataMIC(LoRaWAN1_1, 1, 2, 3, test.FNwkSIntKey, test.SNwkSIntKey)
					case UnconfirmedDataDown, ConfirmedDataDown:
						mic, err = phy.calculateDownlinkDataMIC(LoRaWAN1_1, 1, test.SNwkSIntKey)
					default:
						t.Fatalf("unexpected MType %s", phy.MHDR.MType)
					}

					fmt.Printf("expected mic: %s (%v)\n", mic, mic[:])
				}
				So(ok, ShouldBeTrue)

				if test.EncryptFOpts {
					So(phy.DecryptFOpts(test.NwkSEncKey), ShouldBeNil)
				} else {
					So(phy.DecodeFOptsToMACCommands(), ShouldBeNil)
				}
				So(phy.DecryptFRMPayload(test.AppSKey), ShouldBeNil)
				if macPL, ok := phy.MACPayload.(*MACPayload); ok {
					macPL.FHDR.FCtrl.fOptsLen = 0
				}
				So(phy, ShouldResemble, test.PHYPayload)

				So(test.PHYPayload.EncryptFRMPayload(test.AppSKey), ShouldBeNil)
				if test.EncryptFOpts {
					So(test.PHYPayload.EncryptFOpts(test.NwkSEncKey), ShouldBeNil)
				}

				switch test.PHYPayload.MHDR.MType {
				case UnconfirmedDataUp, ConfirmedDataUp:
					err = test.PHYPayload.SetUplinkDataMIC(LoRaWAN1_1, 1, 2, 3, test.FNwkSIntKey, test.SNwkSIntKey)
				case UnconfirmedDataDown, ConfirmedDataDown:
					err = test.PHYPayload.SetDownlinkDataMIC(LoRaWAN1_1, 1, test.SNwkSIntKey)
				default:
					t.Fatalf("unexpected MType %s", test.PHYPayload.MHDR.MType)
				}
				So(err, ShouldBeNil)

				b, err := test.PHYPayload.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.Bytes)
			})
		}
	})
}

func TestPHYPayloadJoinRequest(t *testing.T) {
	Convey("Given a set of known and an empty PHYPayload", t, func() {
		data, err := base64.StdEncoding.DecodeString("AAQDAgEEAwIBBQQDAgUEAwItEGqZDhI=")
		So(err, ShouldBeNil)

		var phy PHYPayload

		Convey("Then UnmarshalBinary does not fail", func() {
			So(phy.UnmarshalBinary(data), ShouldBeNil)

			Convey("Then the MHDR contains the expected data", func() {
				So(phy.MHDR.MType, ShouldEqual, JoinRequest)
				So(phy.MHDR.Major, ShouldEqual, LoRaWANR1)
			})

			Convey("Then the MIC is valid", func() {
				appKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
				valid, err := phy.ValidateUplinkJoinMIC(appKey)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("Then the MACPayload is of type *JoinRequestPayload", func() {
				jrPl, ok := phy.MACPayload.(*JoinRequestPayload)
				So(ok, ShouldBeTrue)

				Convey("Then the JoinRequestPayload contains the expected data", func() {
					So(jrPl.JoinEUI, ShouldResemble, EUI64{1, 2, 3, 4, 1, 2, 3, 4})
					So(jrPl.DevEUI, ShouldResemble, EUI64{2, 3, 4, 5, 2, 3, 4, 5})
					So(jrPl.DevNonce, ShouldEqual, 4141)
				})
			})

			Convey("When marshalling the PHYPayload", func() {
				b, err := phy.MarshalBinary()
				So(err, ShouldBeNil)

				Convey("Then it equals to the input data", func() {
					So(b, ShouldResemble, data)
				})
			})
		})
	})
}

func TestPHYPayloadJoinAccept(t *testing.T) {
	Convey("Given an empty PHYPayload with empty JoinAcceptPayload", t, func() {
		p := PHYPayload{MACPayload: &JoinAcceptPayload{}}

		Convey("Given an encrypted JoinAccept payload 20493eeb51fba2116f810edb3742975142", func() {
			jaBytes, err := hex.DecodeString("20493eeb51fba2116f810edb3742975142")
			So(err, ShouldBeNil)

			Convey("Then UnmarshalBinary does not return an error", func() {
				So(p.UnmarshalBinary(jaBytes), ShouldBeNil)

				Convey("Given the AppKey 00112233445566778899aabbccddeeff", func() {
					var appKey AES128Key
					appKeyBytes, err := hex.DecodeString("00112233445566778899aabbccddeeff")
					So(err, ShouldBeNil)
					So(appKeyBytes, ShouldHaveLength, 16)
					copy(appKey[:], appKeyBytes)

					Convey("Then decrypting does not return an error", func() {
						So(p.DecryptJoinAcceptPayload(appKey), ShouldBeNil)

						Convey("Then the MACPayload is of type *JoinAcceptPayload", func() {
							jaPL, ok := p.MACPayload.(*JoinAcceptPayload)
							So(ok, ShouldBeTrue)

							Convey("Then the JoinNonce is 5704647", func() {
								So(jaPL.JoinNonce, ShouldEqual, JoinNonce(5704647))
							})

							Convey("Then the NetID is [3]byte{34, 17, 1}", func() {
								So(jaPL.HomeNetID, ShouldEqual, NetID{34, 17, 1})
							})

							Convey("Then the DevAddr is [4]byte{2, 3, 25, 128}", func() {
								So([4]byte(jaPL.DevAddr), ShouldEqual, [4]byte{2, 3, 25, 128})
							})

							Convey("Then the DLSettings is empty", func() {
								So(jaPL.DLSettings, ShouldResemble, DLSettings{})
							})

							Convey("Then the RXDelay = 0", func() {
								So(jaPL.RXDelay, ShouldEqual, 0)
							})

						})

						Convey("Then the MIC is [4]byte{67, 72, 91, 188}", func() {
							So(p.MIC, ShouldEqual, MIC{67, 72, 91, 188})
						})

						Convey("Then the MIC is valid", func() {
							ok, err := p.ValidateDownlinkJoinMIC(JoinRequestType, EUI64{}, 0, appKey)
							So(err, ShouldBeNil)
							So(ok, ShouldBeTrue)
						})
					})
				})
			})
		})

		Convey("Given a JoinAccept with JoinNonce=5704647, NetID=[3]byte{34, 17, 1}, DevAddr=[4]byte{2, 3, 25, 128}", func() {
			p.MHDR = MHDR{
				MType: JoinAccept,
				Major: LoRaWANR1,
			}
			p.MACPayload = &JoinAcceptPayload{
				JoinNonce: 5704647,
				HomeNetID: [3]byte{34, 17, 1},
				DevAddr:   [4]byte{2, 3, 25, 128},
			}

			Convey("Given the AppKey 00112233445566778899aabbccddeeff", func() {
				var appKey AES128Key
				appKeyBytes, err := hex.DecodeString("00112233445566778899aabbccddeeff")
				So(err, ShouldBeNil)
				So(appKeyBytes, ShouldHaveLength, 16)
				copy(appKey[:], appKeyBytes)

				Convey("Then SetMIC does not fail", func() {
					So(p.SetDownlinkJoinMIC(JoinRequestType, EUI64{}, 0, appKey), ShouldBeNil)

					Convey("Then the MIC is [4]byte{67, 72, 91, 188}", func() {
						So(p.MIC, ShouldEqual, MIC{67, 72, 91, 188})
					})

					Convey("Then encrypting does not fail", func() {
						So(p.EncryptJoinAcceptPayload(appKey), ShouldBeNil)

						Convey("Then the hex representation of the packet is 20493eeb51fba2116f810edb3742975142", func() {
							b, err := p.MarshalBinary()
							So(err, ShouldBeNil)
							So(hex.EncodeToString(b), ShouldEqual, "20493eeb51fba2116f810edb3742975142")
						})
					})
				})
			})
		})
	})
}

func TestPHYPayloadRejoinRequest02(t *testing.T) {
	Convey("Given a PHYPayload and a key", t, func() {
		phy := PHYPayload{
			MHDR: MHDR{
				MType: RejoinRequest,
				Major: LoRaWANR1,
			},
			MACPayload: &RejoinRequestType02Payload{
				RejoinType: 2,
				NetID:      NetID{1, 2, 3},
				DevEUI:     EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				RJCount0:   219,
			},
		}
		var key AES128Key

		Convey("Then SetMIC sets the expected MIC", func() {
			So(phy.SetUplinkJoinMIC(key), ShouldBeNil)
			So(phy.MIC, ShouldEqual, MIC{60, 134, 66, 174})
			valid, err := phy.ValidateUplinkJoinMIC(key)
			So(err, ShouldBeNil)
			So(valid, ShouldBeTrue)

			Convey("Then MarshalBinary returns the expected value", func() {
				b, err := phy.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{192, 2, 3, 2, 1, 8, 7, 6, 5, 4, 3, 2, 1, 219, 0, 60, 134, 66, 174})

				Convey("Then UnmarshalBinary returns the expected object", func() {
					var newPhy PHYPayload
					So(newPhy.UnmarshalBinary(b), ShouldBeNil)
					So(newPhy, ShouldResemble, phy)
				})
			})
		})
	})
}

func TestPHYPayloadRejoinRequest1(t *testing.T) {
	Convey("Given a PHYPayload and a key", t, func() {
		phy := PHYPayload{
			MHDR: MHDR{
				MType: RejoinRequest,
				Major: LoRaWANR1,
			},
			MACPayload: &RejoinRequestType1Payload{
				RejoinType: 1,
				JoinEUI:    EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
				RJCount1:   219,
			},
		}
		var key AES128Key

		Convey("Then SetMIC sets the expected MIC", func() {
			So(phy.SetUplinkJoinMIC(key), ShouldBeNil)
			So(phy.MIC, ShouldEqual, MIC{234, 195, 16, 114})

			Convey("Then MarshalBinary returns the expected value", func() {
				b, err := phy.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{192, 1, 8, 7, 6, 5, 4, 3, 2, 1, 16, 15, 14, 13, 12, 11, 10, 9, 219, 0, 234, 195, 16, 114})

				Convey("Then UnmarshalBinary returns the expected object", func() {
					var newPhy PHYPayload
					So(newPhy.UnmarshalBinary(b), ShouldBeNil)
					So(newPhy, ShouldResemble, phy)
				})
			})
		})
	})
}

func ExamplePHYPayload_lorawan10Encode() {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	fPort := uint8(10)

	phy := PHYPayload{
		MHDR: MHDR{
			MType: ConfirmedDataUp,
			Major: LoRaWANR1,
		},
		MACPayload: &MACPayload{
			FHDR: FHDR{
				DevAddr: DevAddr([4]byte{1, 2, 3, 4}),
				FCtrl: FCtrl{
					ADR:       false,
					ADRACKReq: false,
					ACK:       false,
				},
				FCnt: 0,
				FOpts: []Payload{
					&MACCommand{
						CID: DevStatusAns,
						Payload: &DevStatusAnsPayload{
							Battery: 115,
							Margin:  7,
						},
					},
				},
			},
			FPort:      &fPort,
			FRMPayload: []Payload{&DataPayload{Bytes: []byte{1, 2, 3, 4}}},
		},
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetUplinkDataMIC(LoRaWAN1_0, 0, 0, 0, nwkSKey, AES128Key{}); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)
	fmt.Println(string(phyJSON))

	// Output:
	// gAQDAgEDAAAGcwcK4mTU9+EX0sA=
	// [128 4 3 2 1 3 0 0 6 115 7 10 226 100 212 247 225 23 210 192]
	// {"mhdr":{"mType":"ConfirmedDataUp","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":[{"cid":"DevStatusReq","payload":{"battery":115,"margin":7}}]},"fPort":10,"frmPayload":[{"bytes":"4mTU9w=="}]},"mic":"e117d2c0"}
}

func ExamplePHYPayload_lorawan10Decode() {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	var phy PHYPayload
	// use use UnmarshalBinary when decoding a byte-slice
	if err := phy.UnmarshalText([]byte("gAQDAgEDAAAGcwcK4mTU9+EX0sA=")); err != nil {
		panic(err)
	}

	ok, err := phy.ValidateUplinkDataMIC(LoRaWAN1_0, 0, 0, 0, nwkSKey, AES128Key{})
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("invalid mic")
	}

	if err := phy.DecodeFOptsToMACCommands(); err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	if err := phy.DecryptFRMPayload(appSKey); err != nil {
		panic(err)
	}
	macPL, ok := phy.MACPayload.(*MACPayload)
	if !ok {
		panic("*MACPayload expected")
	}

	pl, ok := macPL.FRMPayload[0].(*DataPayload)
	if !ok {
		panic("*DataPayload expected")
	}

	fmt.Println(string(phyJSON))
	fmt.Println(pl.Bytes)

	// Output:
	// {"mhdr":{"mType":"ConfirmedDataUp","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":[{"cid":"DevStatusReq","payload":{"battery":115,"margin":7}}]},"fPort":10,"frmPayload":[{"bytes":"4mTU9w=="}]},"mic":"e117d2c0"}
	// [1 2 3 4]
}

func ExamplePHYPayload_lorawan11EncryptedFoptsEncode() {
	sNwkSIntKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	nwkSEncKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	var fport1 uint8 = 1

	phy := PHYPayload{
		MHDR: MHDR{
			MType: UnconfirmedDataDown,
			Major: LoRaWANR1,
		},
		MACPayload: &MACPayload{
			FHDR: FHDR{
				DevAddr: DevAddr{1, 2, 3, 4},
				FOpts: []Payload{
					&MACCommand{
						CID: LinkCheckAns,
						Payload: &LinkCheckAnsPayload{
							Margin: 7,
							GwCnt:  1,
						},
					},
				},
			},
			FPort: &fport1,
			FRMPayload: []Payload{
				&DataPayload{Bytes: []byte{1, 2, 3, 4}},
			},
		},
	}

	if err := phy.EncryptFOpts(nwkSEncKey); err != nil {
		panic(err)
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetDownlinkDataMIC(LoRaWAN1_1, 0, sNwkSIntKey); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)
	fmt.Println(string(phyJSON))

	// Output:
	// YAQDAgEDAAAirAoB8LRo3ape0To=
	// [96 4 3 2 1 3 0 0 34 172 10 1 240 180 104 221 170 94 209 58]
	// {"mhdr":{"mType":"UnconfirmedDataDown","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":[{"bytes":"IqwK"}]},"fPort":1,"frmPayload":[{"bytes":"8LRo3Q=="}]},"mic":"aa5ed13a"}
}

func ExamplePHYPayload_lorawan11EncryptedFoptsDecode() {
	sNwkSIntKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	nwkSEncKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	var phy PHYPayload
	if err := phy.UnmarshalText([]byte("YAQDAgEDAAAirAoB8LRo3ape0To=")); err != nil {
		panic(err)
	}

	ok, err := phy.ValidateDownlinkDataMIC(LoRaWAN1_1, 0, sNwkSIntKey)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("invalid mic")
	}

	if err := phy.DecryptFOpts(nwkSEncKey); err != nil {
		panic(err)
	}

	if err := phy.DecryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(phyJSON))

	// Output:
	// {"mhdr":{"mType":"UnconfirmedDataDown","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":[{"cid":"LinkCheckReq","payload":{"margin":7,"gwCnt":1}}]},"fPort":1,"frmPayload":[{"bytes":"AQIDBA=="}]},"mic":"aa5ed13a"}
}

func ExamplePHYPayload_proprietaryEncode() {
	phy := PHYPayload{
		MHDR: MHDR{
			MType: Proprietary,
			Major: LoRaWANR1,
		},
		MACPayload: &DataPayload{Bytes: []byte{5, 6, 7, 8, 9, 10}},
		MIC:        MIC{1, 2, 3, 4},
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)
	fmt.Println(string(phyJSON))

	// Output:
	// 4AUGBwgJCgECAwQ=
	// [224 5 6 7 8 9 10 1 2 3 4]
	// {"mhdr":{"mType":"Proprietary","major":"LoRaWANR1"},"macPayload":{"bytes":"BQYHCAkK"},"mic":"01020304"}
}

func ExamplePHYPayload_proprietaryDecode() {
	var phy PHYPayload

	if err := phy.UnmarshalText([]byte("4AUGBwgJCgECAwQ=")); err != nil {
		panic(err)
	}

	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}

	pl, ok := phy.MACPayload.(*DataPayload)
	if !ok {
		panic("*DataPayload expected")
	}

	fmt.Println(phy.MIC)
	fmt.Println(pl.Bytes)
	fmt.Println(string(phyJSON))

	// Output:
	// 01020304
	// [5 6 7 8 9 10]
	// {"mhdr":{"mType":"Proprietary","major":"LoRaWANR1"},"macPayload":{"bytes":"BQYHCAkK"},"mic":"01020304"}
}

func ExamplePHYPayload_joinRequest() {
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	phy := PHYPayload{
		MHDR: MHDR{
			MType: JoinRequest,
			Major: LoRaWANR1,
		},
		MACPayload: &JoinRequestPayload{
			JoinEUI:  [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: 771,
		},
	}

	if err := phy.SetUplinkJoinMIC(appKey); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)

	// Output:
	// AAEBAQEBAQEBAgICAgICAgIDAwm5ezI=
	// [0 1 1 1 1 1 1 1 1 2 2 2 2 2 2 2 2 3 3 9 185 123 50]
}

func ExamplePHYPayload_joinAcceptSend() {
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	joinEUI := EUI64{8, 7, 6, 5, 4, 3, 2, 1}
	devNonce := DevNonce(258)

	phy := PHYPayload{
		MHDR: MHDR{
			MType: JoinAccept,
			Major: LoRaWANR1,
		},
		MACPayload: &JoinAcceptPayload{
			JoinNonce:  65793,
			HomeNetID:  [3]byte{2, 2, 2},
			DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
			DLSettings: DLSettings{RX2DataRate: 0, RX1DROffset: 0},
			RXDelay:    0,
		},
	}

	// set the MIC before encryption
	if err := phy.SetDownlinkJoinMIC(JoinRequestType, joinEUI, devNonce, appKey); err != nil {
		panic(err)
	}
	if err := phy.EncryptJoinAcceptPayload(appKey); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)

	// Output:
	// ICPPM1SJquMYPAvguqje5fM=
	// [32 35 207 51 84 137 170 227 24 60 11 224 186 168 222 229 243]
}

func ExamplePHYPayload_lorawan11JoinAcceptSend() {
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	joinEUI := EUI64{8, 7, 6, 5, 4, 3, 2, 1}
	devNonce := DevNonce(258)

	// note: the DLSettings OptNeg is set to true!
	phy := PHYPayload{
		MHDR: MHDR{
			MType: JoinAccept,
			Major: LoRaWANR1,
		},
		MACPayload: &JoinAcceptPayload{
			JoinNonce:  65793,
			HomeNetID:  [3]byte{2, 2, 2},
			DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
			DLSettings: DLSettings{RX2DataRate: 0, RX1DROffset: 0, OptNeg: true},
			RXDelay:    0,
		},
	}

	// set the MIC before encryption
	if err := phy.SetDownlinkJoinMIC(JoinRequestType, joinEUI, devNonce, appKey); err != nil {
		panic(err)
	}
	if err := phy.EncryptJoinAcceptPayload(appKey); err != nil {
		panic(err)
	}

	str, err := phy.MarshalText()
	if err != nil {
		panic(err)
	}

	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))
	fmt.Println(bytes)

	// Output:
	// IHq+6gawKSDxHALQNI/PGBU=
	// [32 122 190 234 6 176 41 32 241 28 2 208 52 143 207 24 21]
}

func ExamplePHYPayload_readJoinRequest() {
	var phy PHYPayload
	if err := phy.UnmarshalText([]byte("AAQDAgEEAwIBBQQDAgUEAwItEGqZDhI=")); err != nil {
		panic(err)
	}

	jrPL, ok := phy.MACPayload.(*JoinRequestPayload)
	if !ok {
		panic("MACPayload must be a *JoinRequestPayload")
	}

	fmt.Println(phy.MHDR.MType)
	fmt.Println(jrPL.JoinEUI)
	fmt.Println(jrPL.DevEUI)
	fmt.Println(jrPL.DevNonce)

	// Output:
	// JoinRequest
	// 0102030401020304
	// 0203040502030405
	// 4141
}
