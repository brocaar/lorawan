package lorawan

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDevAddr(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		tests := []struct {
			Name      string
			DevAddr   DevAddr
			NetIDType int
			NwkID     []byte
			Bytes     []byte
			String    string
		}{
			{
				Name:      "NetID type 0",
				DevAddr:   DevAddr{91, 255, 255, 255},
				NetIDType: 0,
				NwkID:     []byte{45},
				Bytes:     []byte{255, 255, 255, 91},
				String:    "5bffffff",
			},
			{
				Name:      "NetID type 1",
				DevAddr:   DevAddr{173, 255, 255, 255},
				NetIDType: 1,
				NwkID:     []byte{45},
				Bytes:     []byte{255, 255, 255, 173},
				String:    "adffffff",
			},
			{
				Name:      "NetID type 2",
				DevAddr:   DevAddr{214, 223, 255, 255},
				NetIDType: 2,
				NwkID:     []byte{1, 109},
				Bytes:     []byte{255, 255, 223, 214},
				String:    "d6dfffff",
			},
			{
				Name:      "NetID type 3",
				DevAddr:   DevAddr{235, 111, 255, 255},
				NetIDType: 3,
				NwkID:     []byte{5, 183},
				Bytes:     []byte{255, 255, 111, 235},
				String:    "eb6fffff",
			},
			{
				Name:      "NetID type 4",
				DevAddr:   DevAddr{245, 182, 255, 255},
				NetIDType: 4,
				NwkID:     []byte{11, 109},
				Bytes:     []byte{255, 255, 182, 245},
				String:    "f5b6ffff",
			},
			{
				Name:      "NetID type 5",
				DevAddr:   DevAddr{250, 219, 127, 255},
				NetIDType: 5,
				NwkID:     []byte{22, 219},
				Bytes:     []byte{255, 127, 219, 250},
				String:    "fadb7fff",
			},
			{
				Name:      "NetID type 6",
				DevAddr:   DevAddr{253, 109, 183, 255},
				NetIDType: 6,
				NwkID:     []byte{91, 109},
				Bytes:     []byte{255, 183, 109, 253},
				String:    "fd6db7ff",
			},
			{
				Name:      "NetID type 7",
				DevAddr:   DevAddr{254, 182, 219, 127},
				NetIDType: 7,
				NwkID:     []byte{1, 109, 182},
				Bytes:     []byte{127, 219, 182, 254},
				String:    "feb6db7f",
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				So(test.DevAddr.NetIDType(), ShouldEqual, test.NetIDType)
				So(test.DevAddr.NwkID(), ShouldResemble, test.NwkID)

				b, err := test.DevAddr.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.Bytes)

				var devAddr DevAddr
				So(devAddr.UnmarshalBinary(test.Bytes), ShouldBeNil)
				So(devAddr, ShouldResemble, test.DevAddr)

				b, err = test.DevAddr.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, test.String)

				So(devAddr.UnmarshalText([]byte(test.String)), ShouldBeNil)
				So(devAddr, ShouldEqual, test.DevAddr)
			})
		}
	})
}

func TestDevAddrSetAddrPrefix(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		tests := []struct {
			Name            string
			BeforeDevAddr   DevAddr
			NetID           []NetID
			ExpectedDevAddr []DevAddr
		}{
			{
				Name:            "Set type 0 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{0, 0, 0}, {0, 0, 63}},
				ExpectedDevAddr: []DevAddr{{1, 255, 255, 255}, {127, 255, 255, 255}},
			},
			{
				Name:            "Set type 1 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{32, 0, 0}, {32, 0, 63}},
				ExpectedDevAddr: []DevAddr{{128, 255, 255, 255}, {191, 255, 255, 255}},
			},
			{
				Name:            "Set type 2 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{64, 0, 0}, {64, 1, 255}},
				ExpectedDevAddr: []DevAddr{{192, 15, 255, 255}, {223, 255, 255, 255}},
			},
			{
				Name:            "Set type 3 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{96, 0, 0}, {111, 255, 255}},
				ExpectedDevAddr: []DevAddr{{224, 1, 255, 255}, {239, 255, 255, 255}},
			},
			{
				Name:            "Set type 4 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{128, 0, 0}, {159, 255, 255}},
				ExpectedDevAddr: []DevAddr{{240, 0, 127, 255}, {247, 255, 255, 255}},
			},
			{
				Name:            "Set type 5 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{160, 0, 0}, {191, 255, 255}},
				ExpectedDevAddr: []DevAddr{{248, 0, 31, 255}, {251, 255, 255, 255}},
			},
			{
				Name:            "Set type 6 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{192, 0, 0}, {223, 255, 255}},
				ExpectedDevAddr: []DevAddr{{252, 0, 3, 255}, {253, 255, 255, 255}},
			},
			{
				Name:            "Set type 7 prefix",
				BeforeDevAddr:   DevAddr{255, 255, 255, 255},
				NetID:           []NetID{{224, 0, 0}, {255, 255, 255}},
				ExpectedDevAddr: []DevAddr{{254, 0, 0, 127}, {254, 255, 255, 255}},
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				for i := range test.NetID {
					So(test.BeforeDevAddr.IsNetID(test.NetID[i]), ShouldBeFalse)
					test.BeforeDevAddr.SetAddrPrefix(test.NetID[i])
					So(test.BeforeDevAddr, ShouldEqual, test.ExpectedDevAddr[i])
					So(test.BeforeDevAddr.IsNetID(test.NetID[i]), ShouldBeTrue)
				}
			})
		}
	})
}

func TestFCtrl(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		testTable := []struct {
			FCtrl         FCtrl
			ExpectedBytes []byte
			ExpectedError error
		}{
			{
				FCtrl:         FCtrl{fOptsLen: 16},
				ExpectedError: errors.New("lorawan: max value of FOptsLen is 15"),
			},
			{
				FCtrl:         FCtrl{ADR: true, ADRACKReq: false, ACK: false, FPending: false, fOptsLen: 2},
				ExpectedBytes: []byte{130},
			},
			{
				FCtrl:         FCtrl{ADR: false, ADRACKReq: true, ACK: false, FPending: false, fOptsLen: 3},
				ExpectedBytes: []byte{67},
			},
			{
				FCtrl:         FCtrl{ADR: false, ADRACKReq: false, ACK: true, FPending: false, fOptsLen: 4},
				ExpectedBytes: []byte{36},
			},
			{
				FCtrl:         FCtrl{ADR: false, ADRACKReq: false, ACK: false, FPending: true, fOptsLen: 5},
				ExpectedBytes: []byte{21},
			},
			{
				FCtrl:         FCtrl{ADR: false, ADRACKReq: false, ACK: false, ClassB: true, fOptsLen: 5},
				ExpectedBytes: []byte{21},
			},
			{
				FCtrl:         FCtrl{ADR: true, ADRACKReq: true, ACK: true, FPending: true, fOptsLen: 6},
				ExpectedBytes: []byte{246},
			},
			{
				FCtrl:         FCtrl{ADR: true, ADRACKReq: true, ACK: true, ClassB: true, fOptsLen: 6},
				ExpectedBytes: []byte{246},
			},
		}

		for i, test := range testTable {
			Convey(fmt.Sprintf("Testing: %+v [%d]", test.FCtrl, i), func() {
				b, err := test.FCtrl.MarshalBinary()
				if test.ExpectedError != nil {
					Convey("Then the expected error is returned", func() {
						So(err, ShouldNotBeNil)
						So(err, ShouldResemble, test.ExpectedError)
					})
					return
				}
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.ExpectedBytes)

				Convey("Then unmarshal and marshal results in the same byteslice", func() {
					var fCtrl FCtrl
					So(fCtrl.UnmarshalBinary(b), ShouldBeNil)
					b2, err := fCtrl.MarshalBinary()
					So(err, ShouldBeNil)
					So(b2, ShouldResemble, b)
				})
			})
		}
	})
}

func TestFHDR(t *testing.T) {
	Convey("Given an empty FHDR", t, func() {
		var h FHDR
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0}", func() {
			b, err := h.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given the FCnt contains a value > 16 bits", func() {
			h.FCnt = 65795

			Convey("Then only the least-significant 16 bits are marshalled", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 3, 1})
			})
		})

		Convey("Given DevAddr=67305985, FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true), Fcnt=5, FOpts=[]MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			h.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			h.FCtrl = FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true}
			h.FCnt = 5
			h.FOpts = []Payload{
				&MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}},
			}
			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9})
			})
		})

		Convey("Given FOpts contains 5 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 5; i++ {
				h.FOpts = append(h.FOpts, &MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does not return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldBeNil)
			})
		})

		Convey("Given FOpts contains 6 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 6; i++ {
				h.FOpts = append(h.FOpts, &MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldResemble, errors.New("lorawan: max number of FOpts bytes is 15"))
			})
		})

		Convey("Given uplink=false and slice []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
			b := []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}
			Convey("Then UnmarshalBinary does not return an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)
				h.FOpts, err = decodeDataPayloadToMACCommands(false, h.FOpts)
				So(err, ShouldBeNil)

				Convey("Then DevAddr=[4]{1, 2, 3, 4}", func() {
					So(h.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})

				Convey("Then FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true, fOptsLen=3)", func() {
					So(h.FCtrl, ShouldResemble, FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true, ClassB: true, fOptsLen: 3})
				})

				Convey("Then len(FOpts)=1", func() {
					So(h.FOpts, ShouldHaveLength, 1)
					Convey("Then CID=LinkCheckAns", func() {
						So(h.FOpts[0].(*MACCommand).CID, ShouldEqual, LinkCheckAns)
					})
				})

				Convey("Then Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9)", func() {
					p, ok := h.FOpts[0].(*MACCommand).Payload.(*LinkCheckAnsPayload)
					So(ok, ShouldBeTrue)
					So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 7, GwCnt: 9})
				})
			})
		})

		Convey("Given uplink=false and slice []byte{4, 3, 2, 1, 181, 5, 0, 2, 7, 9, 78, 79} (one known mac-command and some unknown data)", func() {
			b := []byte{4, 3, 2, 1, 181, 5, 0, 2, 7, 9, 78, 79}
			var logBytes bytes.Buffer
			log.SetOutput(&logBytes)

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)
				h.FOpts, err = decodeDataPayloadToMACCommands(false, h.FOpts)
				So(err, ShouldBeNil)

				Convey("Then DevAddr=[4]{1, 2, 3, 4}", func() {
					So(h.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})

				Convey("Then FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true, fOptsLen=5)", func() {
					So(h.FCtrl, ShouldResemble, FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true, ClassB: true, fOptsLen: 5})
				})

				Convey("Then len(FOpts)=3", func() {
					So(h.FOpts, ShouldHaveLength, 3)
					Convey("Then CID=LinkCheckAns", func() {
						So(h.FOpts[0].(*MACCommand).CID, ShouldEqual, LinkCheckAns)
					})

				})

				Convey("Then the remaining mac data is still available", func() {
					So(h.FOpts[1].(*MACCommand).CID, ShouldEqual, 78)
					So(h.FOpts[2].(*MACCommand).CID, ShouldEqual, 79)
				})

				Convey("Then Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9)", func() {
					p, ok := h.FOpts[0].(*MACCommand).Payload.(*LinkCheckAnsPayload)
					So(ok, ShouldBeTrue)
					So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 7, GwCnt: 9})
				})
			})
		})

		Convey("Given uplink=false and slice []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}", func() {
			b := []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := h.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)
				h.FOpts, err = decodeDataPayloadToMACCommands(false, h.FOpts)
				So(err, ShouldResemble, errors.New("lorawan: not enough remaining bytes"))
			})
		})

		Convey("Given FOpts with a MACCommand with non-empty NewChannelReqPayload", func() {
			m := MACCommand{
				CID: NewChannelReq,
				Payload: &NewChannelReqPayload{
					ChIndex: 2,
					Freq:    868100000,
					MaxDR:   5,
					MinDR:   1,
				},
			}
			h.FOpts = []Payload{&m}
			Convey("When it is transformed into binary", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				Convey("Then it can be converted back to the original payload", func() {
					actual := FHDR{}
					So(actual.UnmarshalBinary(false, b), ShouldBeNil)
					actual.FOpts, err = decodeDataPayloadToMACCommands(false, actual.FOpts)
					So(err, ShouldBeNil)
					So(actual.FOpts, ShouldResemble, []Payload{&m})
				})
			})
		})
	})
}
