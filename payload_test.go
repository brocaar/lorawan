package lorawan

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestEUI64(t *testing.T) {
	Convey("Given an empty EUI64", t, func() {
		var eui EUI64

		Convey("When the value is [8]{1, 2, 3, 4, 5, 6, 7, 8}", func() {
			eui = [8]byte{1, 2, 3, 4, 5, 6, 7, 8}

			Convey("Then MarshalText returns 0102030405060708", func() {
				b, err := eui.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "0102030405060708")
			})

			Convey("Then Value returns the expected value", func() {
				v, err := eui.Value()
				So(err, ShouldBeNil)
				So(v, ShouldResemble, driver.Value(eui[:]))
			})
		})

		Convey("Given the string 0102030405060708", func() {
			str := "0102030405060708"

			Convey("Then UnmarshalText returns EUI64{1, 2, 3, 4, 5, 6, 7, 8}", func() {
				err := eui.UnmarshalText([]byte(str))
				So(err, ShouldBeNil)
				So(eui, ShouldResemble, EUI64{1, 2, 3, 4, 5, 6, 7, 8})
			})
		})

		Convey("Given []byte{1, 2, 3, 4, 5, 6, 7, 8}", func() {
			b := []byte{1, 2, 3, 4, 5, 6, 7, 8}
			Convey("Then Scan scans the value correctly", func() {
				So(eui.Scan(b), ShouldBeNil)
				So(eui[:], ShouldResemble, b)
			})
		})
	})
}

func TestDevNonce(t *testing.T) {
	Convey("Given an empty DevNonce", t, func() {
		var nonce DevNonce

		Convey("When setting the dev-nonce", func() {
			nonce = DevNonce(272)

			Convey("Then MarshalBinary returns the expected value", func() {
				b, err := nonce.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{16, 1})
			})
		})

		Convey("Then UnmarshalBinary returns the expected nonce", func() {
			So(nonce.UnmarshalBinary([]byte{16, 1}), ShouldBeNil)
			So(nonce, ShouldEqual, DevNonce(272))
		})
	})
}

func TestJoinNonce(t *testing.T) {
	Convey("Given an empty JoinNonce", t, func() {
		var nonce JoinNonce

		Convey("When setting the app-nonce", func() {
			nonce = JoinNonce(66051)

			Convey("Then MarshalBinary returns the expected value", func() {
				b, err := nonce.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{3, 2, 1})
			})
		})

		Convey("Then UnmarshalBinary returns the expected value", func() {
			So(nonce.UnmarshalBinary([]byte{3, 2, 1}), ShouldBeNil)
			So(nonce, ShouldEqual, 66051)
		})
	})
}

func TestDataPayload(t *testing.T) {
	Convey("Given an empty DataPayload", t, func() {
		var p DataPayload
		Convey("Then MarshalBinary returns []byte{}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldHaveLength, 0)
		})

		Convey("Given Bytes=[]byte{1, 2, 3, 4}", func() {
			p.Bytes = []byte{1, 2, 3, 4}
			Convey("Then MarshalBinary returns []byte{1, 2, 3, 4}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 2, 3, 4})
			})
		})

		Convey("Given the slice []byte{1, 2, 3, 4}", func() {
			b := []byte{1, 2, 3, 4}
			Convey("Then UnmarshalBinary returns DataPayload with Bytes=[]byte{1, 2, 3, 4}", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)
				So(p.Bytes, ShouldNotEqual, b) // make sure we get a new copy!
				So(p.Bytes, ShouldResemble, b)
			})
		})
	})
}

func TestJoinRequestPayload(t *testing.T) {
	Convey("Given an empty JoinRequestPayload", t, func() {
		var p JoinRequestPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given JoinEUI=[8]byte{1, 1, 1, 1, 1, 1, 1, 1}, DevEUI=[8]byte{2, 2, 2, 2, 2, 2, 2, 2} and DevNonce=[2]byte{3, 3}", func() {
			p.JoinEUI = [8]byte{1, 1, 1, 1, 1, 1, 1, 1}
			p.DevEUI = [8]byte{2, 2, 2, 2, 2, 2, 2, 2}
			p.DevNonce = 771
			Convey("Then MarshalBinary returns []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3})
			})
		})

		Convey("Given a slice of bytes with an invalid size", func() {
			b := make([]byte, 17)
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldResemble, errors.New("lorawan: 18 bytes of data are expected"))
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}", func() {
			b := []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}
			Convey("Then UnmarshalBinary returns a JoinRequestPayload with JoinEUI=[8]byte{1, 1, 1, 1, 1, 1, 1, 1}, DevEUI=[8]byte{2, 2, 2, 2, 2, 2, 2, 2} and DevNonce=[2]byte{3, 3}", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, JoinRequestPayload{
					JoinEUI:  [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
					DevEUI:   [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
					DevNonce: 771,
				})
			})
		})
	})
}

func TestCFList(t *testing.T) {
	Convey("Given a test-set", t, func() {
		tests := []struct {
			Name          string
			CFList        CFList
			Bytes         []byte
			ExpectedError error
		}{
			{
				Name: "Invalid channel-frequency",
				CFList: CFList{
					CFListType: CFListChannel,
					Payload: &CFListChannelPayload{
						Channels: [5]uint32{
							868100001,
						},
					},
				},
				ExpectedError: errors.New("lorawan: frequency must be a multiple of 100"),
			},
			{
				Name: "Channel-frequency list",
				CFList: CFList{
					CFListType: CFListChannel,
					Payload: &CFListChannelPayload{
						Channels: [5]uint32{
							867100000,
							867300000,
							867500000,
							867700000,
							867900000,
						},
					},
				},
				Bytes: []byte{24, 79, 132, 232, 86, 132, 184, 94, 132, 136, 102, 132, 88, 110, 132, 0},
			},
			{
				Name: "Channel-mask list (first 8)",
				CFList: CFList{
					CFListType: CFListChannelMask,
					Payload: &CFListChannelMaskPayload{
						ChannelMasks: []ChMask{
							{
								true,
								true,
								true,
								true,
								true,
								true,
								true,
								true,
							},
						},
					},
				},
				Bytes: []byte{255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
			{
				Name: "Channel-mask list (second 8)",
				CFList: CFList{
					CFListType: CFListChannelMask,
					Payload: &CFListChannelMaskPayload{
						ChannelMasks: []ChMask{
							{},
							{
								true,
								true,
								true,
								true,
								true,
								true,
								true,
								true,
							},
						},
					},
				},
				Bytes: []byte{0, 0, 255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				b, err := test.CFList.MarshalBinary()
				if test.ExpectedError != nil {
					So(err, ShouldResemble, test.ExpectedError)
					return
				}
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.Bytes)

				var cFList CFList
				So(cFList.UnmarshalBinary(b), ShouldBeNil)
				So(cFList, ShouldResemble, test.CFList)
			})
		}
	})
}

func TestJoinAcceptPayload(t *testing.T) {
	Convey("Given an empty JoinAcceptPayload", t, func() {
		var p JoinAcceptPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given JoinNonce=65793, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9", func() {
			p.JoinNonce = 65793
			p.HomeNetID = [3]byte{2, 2, 2}
			p.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			p.DLSettings.RX2DataRate = 7
			p.DLSettings.RX1DROffset = 6
			p.RXDelay = 9

			Convey("Then MarshalBinary returns []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9})
			})
		})

		Convey("Given JoinNonce=65793, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9, CFList=867.1, 867.3, 867.5, 867.7, 867.9", func() {
			p.JoinNonce = 65793
			p.HomeNetID = [3]byte{2, 2, 2}
			p.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			p.DLSettings.RX2DataRate = 7
			p.DLSettings.RX1DROffset = 6
			p.RXDelay = 9
			p.CFList = &CFList{
				CFListType: CFListChannel,
				Payload: &CFListChannelPayload{
					Channels: [5]uint32{
						867100000,
						867300000,
						867500000,
						867700000,
						867900000,
					},
				},
			}

			Convey("Then MarshalBinary returns []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9, 24, 79, 132, 232, 86, 132, 184, 94, 132, 136, 102, 132, 88, 110, 132, 0}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9, 24, 79, 132, 232, 86, 132, 184, 94, 132, 136, 102, 132, 88, 110, 132, 0})
			})

		})

		Convey("Given a slice of bytes with an invalid size", func() {
			b := make([]byte, 11)
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldResemble, errors.New("lorawan: 12 or 28 bytes of data are expected (28 bytes if CFList is present)"))
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}", func() {
			b := []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}
			Convey("Then UnmarshalBinary returns a JoinAcceptPayload with JoinNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				So(p.JoinNonce, ShouldEqual, JoinNonce(65793))
				So(p.HomeNetID, ShouldResemble, NetID{2, 2, 2})
				So(p.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				So(p.DLSettings, ShouldResemble, DLSettings{RX2DataRate: 7, RX1DROffset: 6})
				So(p.RXDelay, ShouldEqual, 9)
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9,24,79,132,232,86,132,184,94,132,136,102,132,88,110,132,0}", func() {
			b := []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9, 24, 79, 132, 232, 86, 132, 184, 94, 132, 136, 102, 132, 88, 110, 132, 0}
			Convey("Then UnmarshalBinary returns a JoinAcceptPayload with JoinNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9, CFlist= 867.1,867.3,867.5,867.7,867.9", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				So(p.JoinNonce, ShouldResemble, JoinNonce(65793))
				So(p.HomeNetID, ShouldResemble, NetID{2, 2, 2})
				So(p.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				So(p.DLSettings, ShouldResemble, DLSettings{RX2DataRate: 7, RX1DROffset: 6})
				So(p.RXDelay, ShouldEqual, 9)
				So(p.CFList, ShouldNotBeNil)
				So(p.CFList, ShouldResemble, &CFList{
					CFListType: CFListChannel,
					Payload: &CFListChannelPayload{
						Channels: [5]uint32{
							867100000,
							867300000,
							867500000,
							867700000,
							867900000,
						},
					},
				})
			})
		})
	})
}

func TestRejoinRequestType02Payload(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		testTable := []struct {
			Payload       RejoinRequestType02Payload
			Bytes         []byte
			ExpectedError error
		}{
			{
				Payload: RejoinRequestType02Payload{
					RejoinType: RejoinRequestType0,
					NetID:      NetID{1, 2, 3},
					DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
					RJCount0:   219,
				},
				Bytes: []byte{0, 3, 2, 1, 16, 15, 14, 13, 12, 11, 10, 9, 219, 0},
			},
			{
				Payload: RejoinRequestType02Payload{
					RejoinType: RejoinRequestType1,
					NetID:      NetID{1, 2, 3},
					DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
					RJCount0:   219,
				},
				ExpectedError: errors.New("lorawan: RejoinType must be 0 or 2"),
			},
			{
				Payload: RejoinRequestType02Payload{
					RejoinType: RejoinRequestType2,
					NetID:      NetID{1, 2, 3},
					DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
					RJCount0:   219,
				},
				Bytes: []byte{2, 3, 2, 1, 16, 15, 14, 13, 12, 11, 10, 9, 219, 0},
			},
		}

		for _, test := range testTable {
			b, err := test.Payload.MarshalBinary()
			So(err, ShouldResemble, test.ExpectedError)
			So(b, ShouldResemble, test.Bytes)

			if test.ExpectedError != nil {
				continue
			}

			var pl RejoinRequestType02Payload
			So(pl.UnmarshalBinary(true, b), ShouldBeNil)
			So(pl, ShouldResemble, test.Payload)
		}
	})
}

func TestRejoinRequestType1Payload(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		testTable := []struct {
			Payload       RejoinRequestType1Payload
			Bytes         []byte
			ExpectedError error
		}{
			{
				Payload: RejoinRequestType1Payload{
					RejoinType: RejoinRequestType1,
					JoinEUI:    EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
					RJCount1:   219,
				},
				Bytes: []byte{1, 8, 7, 6, 5, 4, 3, 2, 1, 16, 15, 14, 13, 12, 11, 10, 9, 219, 0},
			},
			{
				Payload: RejoinRequestType1Payload{
					RejoinType: RejoinRequestType2,
					JoinEUI:    EUI64{1, 2, 3, 4, 5, 6, 7, 8},
					DevEUI:     EUI64{9, 10, 11, 12, 13, 14, 15, 16},
					RJCount1:   219,
				},
				ExpectedError: errors.New("lorawan: RejoinType must be 1"),
			},
		}

		for _, test := range testTable {
			b, err := test.Payload.MarshalBinary()
			So(err, ShouldResemble, test.ExpectedError)
			So(b, ShouldResemble, test.Bytes)

			if test.ExpectedError != nil {
				continue
			}

			var pl RejoinRequestType1Payload
			So(pl.UnmarshalBinary(true, b), ShouldBeNil)
			So(pl, ShouldResemble, test.Payload)
		}
	})
}
