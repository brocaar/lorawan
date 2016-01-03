package lorawan

import (
	"encoding/base64"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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

		phy := NewPHYPayload(true) // uplink=true

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
					So(macPl.FPort, ShouldEqual, 1)
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
					So(macPl.DecryptFRMPayload(appSKey), ShouldBeNil)

					Convey("Then the DataPayload contains the expected data", func() {
						So(macPl.FRMPayload, ShouldHaveLength, 1)
						dataPl, ok := macPl.FRMPayload[0].(*DataPayload)
						So(ok, ShouldBeTrue)
						So(dataPl.Bytes, ShouldResemble, []byte("hello"))
					})

					Convey("When encrypting the FRMPayload again and marshalling the PHYPayload", func() {
						So(macPl.EncryptFRMPayload(appSKey), ShouldBeNil)
						b, err := phy.MarshalBinary()
						So(err, ShouldBeNil)

						Convey("Then it equals to the input data", func() {
							So(b, ShouldResemble, data)
						})
					})
				})
			})

			Convey("When using GobEncode", func() {
				b, err := phy.GobEncode()
				So(err, ShouldBeNil)

				Convey("Then when using GobDecode, the packet is uplink and the MIC valid", func() {
					newPHY := &PHYPayload{}
					So(newPHY.GobDecode(b), ShouldBeNil)
					So(newPHY.uplink, ShouldBeTrue)

					valid, err := newPHY.ValidateMIC([16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})
					So(err, ShouldBeNil)
					So(valid, ShouldBeTrue)
				})
			})
		})
	})
}

func TestPHYPayloadJoinRequest(t *testing.T) {
	Convey("Given a set of known and an empty PHYPayload", t, func() {
		data, err := base64.StdEncoding.DecodeString("AAQDAgEEAwIBBQQDAgUEAwItEGqZDhI=")
		So(err, ShouldBeNil)

		phy := NewPHYPayload(true) // uplink=true

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
					So(jrPl.DevNonce, ShouldResemble, [2]byte{16, 45})
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
	// todo: disabled for now, need some real-world testing to make sure the expected data is correct
	SkipConvey("Given an empty PHYPayload with empty JoinAcceptPayload", t, func() {
		p := PHYPayload{MACPayload: &JoinAcceptPayload{}}
		Convey("Then MarshalBinary returns []byte with 17 0x00", func() {
			exp := make([]byte, 17)
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, exp)
		})

		Convey("Given MHDR=(MType=JoinAccept, Major=LoRaWANR1), MACPayload=JoinAcceptPayload(AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=[4]byte{1, 2, 3, 4}, DLSettings=(RX2DataRate=1, RX1DRoffset=2), RXDelay=7), MIC=[4]byte{8, 9 , 10, 11}", func() {
			p.MHDR = MHDR{MType: JoinAccept, Major: LoRaWANR1}
			p.MACPayload = &JoinAcceptPayload{
				AppNonce:   [3]byte{1, 1, 1},
				NetID:      [3]byte{2, 2, 2},
				DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
				DLSettings: DLsettings{RX2DataRate: 1, RX1DRoffset: 2},
				RXDelay:    7,
			}
			p.MIC = [4]byte{8, 9, 10, 11}

			// no encryption and invalid MIC
			Convey("Then MarshalBinary returns []byte{32, 1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 33, 7, 8, 9, 10, 11}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{32, 1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 33, 7, 8, 9, 10, 11})
			})

			Convey("Given AppKey [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}", func() {
				appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

				Convey("Then ValidateMIC returns false", func() {
					v, err := p.ValidateMIC(appKey)
					So(err, ShouldBeNil)
					So(v, ShouldBeFalse)
				})

				Convey("Given SetMIC is called", func() {
					err := p.SetMIC(appKey)
					So(err, ShouldBeNil)

					// todo: validate if this mic is actually valid
					Convey("Then SetMIC sets the MIC to [4]byte{27, 61, 112, 162}", func() {
						So(p.MIC, ShouldResemble, [4]byte{27, 61, 112, 162})
					})

					Convey("Given EncryptMACPayload is called", func() {
						err := p.EncryptMACPayload(appKey)
						So(err, ShouldBeNil)

						Convey("Then MACPayload should be of type *DataPayload", func() {
							dp, ok := p.MACPayload.(*DataPayload)
							So(ok, ShouldBeTrue)

							Convey("Then DataPayload Bytes equals []byte{234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205}", func() {
								So(dp.Bytes, ShouldResemble, []byte{234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205})
							})

							Convey("Then marshalBinary returns []byte{32, 234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205, 27, 61, 112, 162}", func() {
								b, err := p.MarshalBinary()
								So(err, ShouldBeNil)
								So(b, ShouldResemble, []byte{32, 234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205, 27, 61, 112, 162})
							})
						})
					})
				})

				Convey("Given the MIC is [4]byte{27, 61, 112, 162}", func() {
					p.MIC = [4]byte{27, 61, 112, 162}
					Convey("Then ValidateMIC returns true", func() {
						v, err := p.ValidateMIC(appKey)
						So(err, ShouldBeNil)
						So(v, ShouldBeTrue)
					})
				})
			})
		})

		Convey("Given the slice []byte{32, 234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205, 27, 61, 112, 162}", func() {
			b := []byte{32, 234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205, 27, 61, 112, 162}

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)

				Convey("Then MHDR=(MType=JoinAccept, Major=LoRaWANR1)", func() {
					So(p.MHDR, ShouldResemble, MHDR{MType: JoinAccept, Major: LoRaWANR1})
				})

				Convey("Then MACPayload is of type *DataPayload", func() {
					dp, ok := p.MACPayload.(*DataPayload)
					So(ok, ShouldBeTrue)

					Convey("Then Bytes equals []byte{234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205}", func() {
						So(dp.Bytes, ShouldResemble, []byte{234, 201, 51, 48, 151, 50, 166, 172, 136, 105, 14, 81, 71, 167, 87, 205})
					})

					Convey("Given AppKey [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}", func() {
						appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

						Convey("Given DecryptMACPayload is called", func() {
							err := p.DecryptMACPayload(appKey)
							So(err, ShouldBeNil)

							Convey("Then MACPayload is of type *JoinAcceptPayload", func() {
								ja, ok := p.MACPayload.(*JoinAcceptPayload)
								So(ok, ShouldBeTrue)

								Convey("Then MACPayload=JoinAcceptPayload(AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=[4]byte{1, 2, 3, 4}, DLSettings=(RX2DataRate=1, RX1DRoffset=2), RXDelay=7", func() {
									So(ja, ShouldResemble, &JoinAcceptPayload{
										AppNonce:   [3]byte{1, 1, 1},
										NetID:      [3]byte{2, 2, 2},
										DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
										DLSettings: DLsettings{RX2DataRate: 1, RX1DRoffset: 2},
										RXDelay:    7,
									})
								})

								Convey("Then ValidateMIC returns true", func() {
									v, err := p.ValidateMIC(appKey)
									So(err, ShouldBeNil)
									So(v, ShouldBeTrue)
								})
							})
						})
					})
				})

				Convey("Then MIC=[4]byte{27, 61, 112, 162}", func() {
					So(p.MIC, ShouldResemble, [4]byte{27, 61, 112, 162})
				})
			})
		})
	})
}

func ExampleNewPHYPayload() {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	// uplink and downlink messages are (un)marshalled and encrypted / decrypted
	// differently
	uplink := true

	macPayload := NewMACPayload(uplink)
	macPayload.FHDR = FHDR{
		DevAddr: DevAddr([4]byte{1, 2, 3, 4}),
		FCtrl: FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt:  0,
		FOpts: []MACCommand{}, // you can leave this out when there is no MAC command to send
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []Payload{&DataPayload{Bytes: []byte{1, 2, 3, 4}}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := NewPHYPayload(uplink)
	payload.MHDR = MHDR{
		MType: ConfirmedDataUp,
		Major: LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	bytes, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes)

	// Output:
	// [128 4 3 2 1 0 0 0 10 226 100 212 247 181 106 14 117]
}

func ExampleNewPHYPayload_joinRequest() {
	uplink := true
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	payload := NewPHYPayload(uplink)
	payload.MHDR = MHDR{
		MType: JoinRequest,
		Major: LoRaWANR1,
	}
	payload.MACPayload = &JoinRequestPayload{
		AppEUI:   [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
		DevEUI:   [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
		DevNonce: [2]byte{3, 3},
	}

	if err := payload.SetMIC(appKey); err != nil {
		panic(err)
	}

	bytes, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes)

	// Output:
	// [0 1 1 1 1 1 1 1 1 2 2 2 2 2 2 2 2 3 3 9 185 123 50]
}

func ExampleNewPHYPayload_joinAcceptSend() {
	uplink := false
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	payload := NewPHYPayload(uplink)
	payload.MHDR = MHDR{
		MType: JoinAccept,
		Major: LoRaWANR1,
	}
	payload.MACPayload = &JoinAcceptPayload{
		AppNonce:   [3]byte{1, 1, 1},
		NetID:      [3]byte{2, 2, 2},
		DevAddr:    DevAddr([4]byte{1, 2, 3, 4}),
		DLSettings: DLsettings{RX2DataRate: 0, RX1DRoffset: 0},
		RXDelay:    0,
	}
	// set the MIC before encryption
	if err := payload.SetMIC(appKey); err != nil {
		panic(err)
	}
	if err := payload.EncryptMACPayload(appKey); err != nil {
		panic(err)
	}

	bytes, err := payload.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes)

	// Output:
	// [32 35 207 51 84 137 170 227 24 60 11 224 186 168 222 229 243 52 73 242 18]
}

func ExampleNewPHYPayload_joinAcceptReceive() {
	uplink := false
	appKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	bytes := []byte{32, 171, 84, 244, 227, 34, 30, 148, 118, 211, 1, 33, 90, 24, 50, 81, 139, 128, 229, 23, 154}

	payload := NewPHYPayload(uplink)
	if err := payload.UnmarshalBinary(bytes); err != nil {
		panic(err)
	}

	if err := payload.DecryptMACPayload(appKey); err != nil {
		panic(err)
	}

	_, ok := payload.MACPayload.(*JoinAcceptPayload)
	if !ok {
		panic("*JoinAcceptPayload expected")
	}

	v, err := payload.ValidateMIC(appKey)
	if err != nil {
		panic(err)
	}

	fmt.Println(v)

	// Output:
	// true
}
