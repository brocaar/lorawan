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

func TestPHYPayloadData(t *testing.T) {
	Convey("Given a set of known data and an empty PHYPayload", t, func() {
		data, err := base64.StdEncoding.DecodeString("QAQDAgGAAQABppRkJhXWw7WC")
		So(err, ShouldBeNil)

		var phy PHYPayload

		Convey("Then UnmarshalBinary does not fail", func() {
			So(phy.UnmarshalBinary(data), ShouldBeNil)

			Convey("Then the MIC is valid", func() {
				nwkSKey := [16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}
				valid, err := phy.ValidateMIC(nwkSKey)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("Then the MHDR contains the expected data", func() {
				So(phy.MHDR.MType, ShouldEqual, UnconfirmedDataUp)
				So(phy.MHDR.Major, ShouldEqual, LoRaWANR1)
			})

			Convey("Then PHYPayload contains a MACPayload", func() {
				macPl, ok := phy.MACPayload.(*MACPayload)
				So(ok, ShouldBeTrue)

				Convey("Then FPort is correct", func() {
					So(macPl.FPort, ShouldNotBeNil)
					So(*macPl.FPort, ShouldEqual, 1)
				})

				Convey("Then FHDR contains the expcted data", func() {
					So(macPl.FHDR, ShouldResemble, FHDR{
						DevAddr: [4]byte{1, 2, 3, 4},
						FCnt:    1,
						FCtrl:   FCtrl{ADR: true},
					})
				})

				Convey("Then decrypting the FRMPayload does not error", func() {
					appSKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
					So(phy.DecryptFRMPayload(appSKey), ShouldBeNil)

					Convey("Then the DataPayload contains the expected data", func() {
						So(macPl.FRMPayload, ShouldHaveLength, 1)
						dataPl, ok := macPl.FRMPayload[0].(*DataPayload)
						So(ok, ShouldBeTrue)
						So(dataPl.Bytes, ShouldResemble, []byte("hello"))
					})

					Convey("When encrypting the FRMPayload again and marshalling the PHYPayload", func() {
						So(phy.EncryptFRMPayload(appSKey), ShouldBeNil)
						b, err := phy.MarshalBinary()
						So(err, ShouldBeNil)

						Convey("Then it equals to the input data", func() {
							So(b, ShouldResemble, data)
						})
					})
				})
			})
		})
	})
}

func TestPHYPayloadMAC(t *testing.T) {
	nwkSKey := [16]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	fPort := uint8(0)

	Convey("Given two MAC commands", t, func() {
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

		Convey("When the MAC commands are added to the FRMPayload", func() {
			phy := PHYPayload{
				MHDR: MHDR{
					MType: UnconfirmedDataUp,
					Major: LoRaWANR1,
				},
				MACPayload: &MACPayload{
					FPort: &fPort,
					FHDR: FHDR{
						DevAddr: [4]byte{1, 2, 3, 4},
					},
					FRMPayload: []Payload{&mac1, &mac2},
				},
			}

			Convey("When marshaling the packet without encrypting", func() {
				b, err := phy.MarshalBinary()
				So(err, ShouldBeNil)

				Convey("And unmarshaling from this slice of bytes", func() {
					So(phy.UnmarshalBinary(b), ShouldBeNil)

					Convey("Then the MAC command is stored as DataPayload", func() {
						macPL, ok := phy.MACPayload.(*MACPayload)
						So(ok, ShouldBeTrue)

						So(macPL.FRMPayload, ShouldHaveLength, 1)
						So(macPL.FRMPayload[0], ShouldHaveSameTypeAs, &DataPayload{})

						Convey("When calling DecodeFRMPayloadToMACCommands", func() {
							So(phy.DecodeFRMPayloadToMACCommands(), ShouldBeNil)

							Convey("The FRMPayload has been decoded as MACCommands", func() {
								So(macPL.FRMPayload, ShouldHaveLength, 2)
								So(macPL.FRMPayload, ShouldResemble, []Payload{&mac1, &mac2})
							})
						})
					})
				})
			})

			Convey("When encrypting the packet", func() {
				So(phy.EncryptFRMPayload(nwkSKey), ShouldBeNil)

				Convey("Then the MIC is as expected", func() {
					So(phy.SetMIC(nwkSKey), ShouldBeNil)
					So(phy.MIC, ShouldResemble, MIC{238, 106, 165, 8})

					Convey("Then the binary slice is as expected", func() {
						b, err := phy.MarshalBinary()
						So(err, ShouldBeNil)
						So(b, ShouldResemble, []byte{64, 4, 3, 2, 1, 0, 0, 0, 0, 105, 54, 158, 238, 106, 165, 8})
					})
				})
			})

			Convey("When unmarshaling a binary slice containing the MAC commands", func() {
				var phy PHYPayload
				b := []byte{64, 4, 3, 2, 1, 0, 0, 0, 0, 105, 54, 158, 238, 106, 165, 8}

				So(phy.UnmarshalBinary(b), ShouldBeNil)

				Convey("Then marshaling the PHYPayload results in the same slice", func() {
					b2, err := phy.MarshalBinary()
					So(err, ShouldBeNil)
					So(b2, ShouldResemble, b)
				})

				Convey("Then the MIC is valid", func() {
					ok, err := phy.ValidateMIC(nwkSKey)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
				})

				Convey("When decrypting", func() {
					So(phy.DecryptFRMPayload(nwkSKey), ShouldBeNil)

					Convey("Then the FRMPayload contains the MAC commands", func() {
						macPL, ok := phy.MACPayload.(*MACPayload)
						So(ok, ShouldBeTrue)
						So(macPL.FRMPayload, ShouldHaveLength, 2)

						So(macPL.FRMPayload, ShouldResemble, []Payload{&mac1, &mac2})
					})
				})
			})
		})

		Convey("When the MAC commands are added to the FOpts", func() {
			fPort = 1

			phy := PHYPayload{
				MHDR: MHDR{
					MType: UnconfirmedDataUp,
					Major: LoRaWANR1,
				},
				MACPayload: &MACPayload{
					FPort: &fPort,
					FHDR: FHDR{
						DevAddr: [4]byte{1, 2, 3, 4},
						FOpts:   []MACCommand{mac1, mac2},
					},
					FRMPayload: []Payload{&DataPayload{Bytes: []byte{1, 2, 3, 4}}},
				},
			}

			Convey("When encrypting the packet", func() {
				So(phy.EncryptFRMPayload(nwkSKey), ShouldBeNil)

				Convey("Then the MIC is as expected", func() {
					So(phy.SetMIC(nwkSKey), ShouldBeNil)
					So(phy.MIC, ShouldResemble, MIC{182, 77, 192, 57})

					Convey("Then the binary slice is as expected", func() {
						b, err := phy.MarshalBinary()
						So(err, ShouldBeNil)

						So(b, ShouldResemble, []byte{64, 4, 3, 2, 1, 3, 0, 0, 2, 3, 5, 1, 106, 55, 152, 245, 182, 77, 192, 57})
					})
				})
			})

			Convey("When unmarshaling a binary slice containg the MAC commands", func() {
				var phy PHYPayload
				b := []byte{64, 4, 3, 2, 1, 3, 0, 0, 2, 3, 5, 1, 106, 55, 152, 245, 182, 77, 192, 57}
				So(phy.UnmarshalBinary(b), ShouldBeNil)

				Convey("Then marshaling it again results in the same slice", func() {
					b2, err := phy.MarshalBinary()
					So(err, ShouldBeNil)
					So(b2, ShouldResemble, b)
				})

				Convey("Then the MIC is valid", func() {
					ok, err := phy.ValidateMIC(nwkSKey)
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
				})

				Convey("Then the FOpts contains the same MAC commands", func() {
					macPL, ok := phy.MACPayload.(*MACPayload)
					So(ok, ShouldBeTrue)
					So(macPL.FHDR.FOpts, ShouldHaveLength, 2)
					So(macPL.FHDR.FOpts, ShouldResemble, []MACCommand{mac1, mac2})
				})
			})
		})
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
				valid, err := phy.ValidateMIC(appKey)
				So(err, ShouldBeNil)
				So(valid, ShouldBeTrue)
			})

			Convey("Then the MACPayload is of type *JoinRequestPayload", func() {
				jrPl, ok := phy.MACPayload.(*JoinRequestPayload)
				So(ok, ShouldBeTrue)

				Convey("Then the JoinRequestPayload contains the expected data", func() {
					So(jrPl.AppEUI, ShouldResemble, EUI64{1, 2, 3, 4, 1, 2, 3, 4})
					So(jrPl.DevEUI, ShouldResemble, EUI64{2, 3, 4, 5, 2, 3, 4, 5})
					So(jrPl.DevNonce, ShouldResemble, DevNonce{16, 45})
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

							Convey("Then the AppNonce is [3]byte{87, 11, 199}", func() {
								So(jaPL.AppNonce, ShouldEqual, AppNonce{87, 11, 199})
							})

							Convey("Then the NetID is [3]byte{34, 17, 1}", func() {
								So(jaPL.NetID, ShouldEqual, NetID{34, 17, 1})
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
							ok, err := p.ValidateMIC(appKey)
							So(err, ShouldBeNil)
							So(ok, ShouldBeTrue)
						})
					})
				})
			})
		})

		Convey("Given a JoinAccept with AppNonce=[3]byte{87, 11, 199}, NetID=[3]byte{34, 17, 1}, DevAddr=[4]byte{2, 3, 25, 128}", func() {
			p.MHDR = MHDR{
				MType: JoinAccept,
				Major: LoRaWANR1,
			}
			p.MACPayload = &JoinAcceptPayload{
				AppNonce: [3]byte{87, 11, 199},
				NetID:    [3]byte{34, 17, 1},
				DevAddr:  [4]byte{2, 3, 25, 128},
			}

			Convey("Given the AppKey 00112233445566778899aabbccddeeff", func() {
				var appKey AES128Key
				appKeyBytes, err := hex.DecodeString("00112233445566778899aabbccddeeff")
				So(err, ShouldBeNil)
				So(appKeyBytes, ShouldHaveLength, 16)
				copy(appKey[:], appKeyBytes)

				Convey("Then SetMIC does not fail", func() {
					So(p.SetMIC(appKey), ShouldBeNil)

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

func ExamplePHYPayload_encode() {
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
				FCnt:  0,
				FOpts: []MACCommand{}, // you can leave this out when there is no MAC command to send
			},
			FPort:      &fPort,
			FRMPayload: []Payload{&DataPayload{Bytes: []byte{1, 2, 3, 4}}},
		},
	}

	if err := phy.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	if err := phy.SetMIC(nwkSKey); err != nil {
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
	// gAQDAgEAAAAK4mTU97VqDnU=
	// [128 4 3 2 1 0 0 0 10 226 100 212 247 181 106 14 117]
	// {"mhdr":{"mType":"ConfirmedDataUp","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":[]},"fPort":10,"frmPayload":[{"bytes":"4mTU9w=="}]},"mic":"b56a0e75"}
}

func ExamplePHYPayload_decode() {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	var phy PHYPayload
	// use use UnmarshalBinary when decoding a byte-slice
	if err := phy.UnmarshalText([]byte("gAQDAgEAAAAK4mTU97VqDnU=")); err != nil {
		panic(err)
	}

	ok, err := phy.ValidateMIC(nwkSKey)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("invalid mic")
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
	// {"mhdr":{"mType":"ConfirmedDataUp","major":"LoRaWANR1"},"macPayload":{"fhdr":{"devAddr":"01020304","fCtrl":{"adr":false,"adrAckReq":false,"ack":false,"fPending":false,"classB":false},"fCnt":0,"fOpts":null},"fPort":10,"frmPayload":[{"bytes":"4mTU9w=="}]},"mic":"b56a0e75"}
	// [1 2 3 4]
}

func ExamplePHYPayload_proprietary_encode() {
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

func ExamplePHYPayload_proprietary_decode() {
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
			AppEUI:   [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			DevEUI:   [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			DevNonce: [2]byte{3, 3},
		},
	}

	if err := phy.SetMIC(appKey); err != nil {
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

	phy := PHYPayload{
		MHDR: MHDR{
			MType: JoinAccept,
			Major: LoRaWANR1,
		},
		MACPayload: &JoinAcceptPayload{
			AppNonce:   [3]byte{1, 1, 1},
			NetID:      [3]byte{2, 2, 2},
			DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
			DLSettings: DLSettings{RX2DataRate: 0, RX1DROffset: 0},
			RXDelay:    0,
		},
	}

	// set the MIC before encryption
	if err := phy.SetMIC(appKey); err != nil {
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
	fmt.Println(jrPL.AppEUI)
	fmt.Println(jrPL.DevEUI)
	fmt.Println(jrPL.DevNonce)

	// Output:
	// JoinRequest
	// 0102030401020304
	// 0203040502030405
	// 102d
}
