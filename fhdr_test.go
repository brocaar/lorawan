package lorawan

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDevAddr(t *testing.T) {
	Convey("Given an empty DevAddr", t, func() {
		var a DevAddr
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0}", func() {
			b, err := a.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0})
		})

		Convey("Given the DevAddr{255, 1, 1, 1}", func() {
			a = DevAddr{255, 1, 1, 1}
			Convey("Then NwkID returns byte(127)", func() {
				So(a.NwkID(), ShouldEqual, byte(127))
			})
		})

		Convey("Given the DevAddr{1, 2, 3, 4}", func() {
			a = DevAddr{1, 2, 3, 4}
			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1}", func() {
				b, err := a.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1})
			})

			Convey("Then MarshalJSON returns \"01020304\"", func() {
				b, err := a.MarshalJSON()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, `"01020304"`)
			})
		})

		Convey("Given the slice []byte{4, 3, 2, 1}", func() {
			b := []byte{4, 3, 2, 1}
			Convey("Then UnmarshalBinary returns DevAddr{1, 2, 3, 4}", func() {
				err := a.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(a, ShouldResemble, DevAddr{1, 2, 3, 4})
			})
		})

		Convey("Given the string \"01020304\"", func() {
			str := `"01020304"`
			Convey("Then UnmarshalJSON returns DevAddr{1, 2, 3, 4}", func() {
				err := a.UnmarshalJSON([]byte(str))
				So(err, ShouldBeNil)
				So(a, ShouldResemble, DevAddr{1, 2, 3, 4})
			})
		})
	})
}

func TestFCtrl(t *testing.T) {
	Convey("Given an empty FCtrl", t, func() {
		var c FCtrl
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := c.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given FOptsLen > 15", func() {
			c.fOptsLen = 16
			Convey("Then MarshalBinary returns an error", func() {
				_, err := c.MarshalBinary()
				So(err, ShouldNotBeNil)
			})
		})

		testTable := []struct {
			ADR       bool
			ADRACKReq bool
			ACK       bool
			FPending  bool
			FOptsLen  uint8
			Bytes     []byte
		}{
			{true, false, false, false, 2, []byte{130}},
			{false, true, false, false, 3, []byte{67}},
			{false, false, true, false, 4, []byte{36}},
			{false, false, false, true, 5, []byte{21}},
			{true, true, true, true, 6, []byte{246}},
		}

		for _, test := range testTable {
			Convey(fmt.Sprintf("Given ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, fOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
				c.ADR = test.ADR
				c.ADRACKReq = test.ADRACKReq
				c.ACK = test.ACK
				c.FPending = test.FPending
				c.fOptsLen = test.FOptsLen
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := c.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given the slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a FCtrl with ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, fOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
					err := c.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(c, ShouldResemble, FCtrl{ADR: test.ADR, ADRACKReq: test.ADRACKReq, ACK: test.ACK, FPending: test.FPending, fOptsLen: test.FOptsLen})
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

		Convey("Given uplink=false, DevAddr=67305985, FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true), Fcnt=5, FOpts=[]MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			h.uplink = false
			h.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			h.FCtrl = FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true}
			h.FCnt = 5
			h.FOpts = []MACCommand{
				{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}},
			}
			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9})
			})
		})

		Convey("Given FOpts contains 5 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 5; i++ {
				h.FOpts = append(h.FOpts, MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does not return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldBeNil)
			})
		})

		Convey("Given FOpts contains 6 times MACCommand{(CID=LinkCheckAns, Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9))}", func() {
			for i := 0; i < 6; i++ {
				h.FOpts = append(h.FOpts, MACCommand{CID: LinkCheckAns, Payload: &LinkCheckAnsPayload{Margin: 7, GwCnt: 9}})
			}
			Convey("Then MarshalBinary does return an error", func() {
				_, err := h.MarshalBinary()
				So(err, ShouldResemble, errors.New("lorawan: max number of FOpts bytes is 15"))
			})
		})

		Convey("Given uplink=false and slice []byte{4, 2, 2, 1, 179, 5, 0, 2, 7, 9}", func() {
			b := []byte{4, 3, 2, 1, 179, 5, 0, 2, 7, 9}
			h.uplink = false
			Convey("Then UnmarshalBinary does not return an error", func() {
				err := h.UnmarshalBinary(b)
				So(err, ShouldBeNil)

				Convey("Then DevAddr=[4]{1, 2, 3, 4}", func() {
					So(h.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})

				Convey("Then FCtrl=FCtrl(ADR=true, ADRACKReq=false, ACK=true, FPending=true, fOptsLen=3)", func() {
					So(h.FCtrl, ShouldResemble, FCtrl{ADR: true, ADRACKReq: false, ACK: true, FPending: true, fOptsLen: 3})
				})

				Convey("Then len(FOpts)=1", func() {
					So(h.FOpts, ShouldHaveLength, 1)
					Convey("Then CID=LinkCheckAns", func() {
						So(h.FOpts[0].CID, ShouldEqual, LinkCheckAns)
					})

				})

				Convey("Then Payload=LinkCheckAnsPayload(Margin=7, GwCnt=9)", func() {
					p, ok := h.FOpts[0].Payload.(*LinkCheckAnsPayload)
					So(ok, ShouldBeTrue)
					So(p, ShouldResemble, &LinkCheckAnsPayload{Margin: 7, GwCnt: 9})
				})
			})
		})

		Convey("Given uplink=false and slice []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}", func() {
			h.uplink = false
			b := []byte{1, 2, 3, 4, 179, 5, 0, 2, 7}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := h.UnmarshalBinary(b)
				So(err, ShouldResemble, errors.New("lorawan: not enough remaining bytes"))
			})
		})
	})
}
