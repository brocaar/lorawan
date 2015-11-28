package lorawan

import (
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

		Convey("Given The DevAddr{1, 2, 3, 4}", func() {
			a = DevAddr{1, 2, 3, 4}
			Convey("Then MarshalBinary returns []byte{1, 2, 3, 4}", func() {
				b, err := a.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 2, 3, 4})
			})
		})

		Convey("Given the slice []byte{1, 2, 3, 4}", func() {
			b := []byte{1, 2, 3, 4}
			Convey("Then UnmarshalBinary returns DevAddr{1, 2, 3, 4}", func() {
				err := a.UnmarshalBinary(b)
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
			c.FOptsLen = 16
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
			Convey(fmt.Sprintf("Given ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, FOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
				c.ADR = test.ADR
				c.ADRACKReq = test.ADRACKReq
				c.ACK = test.ACK
				c.FPending = test.FPending
				c.FOptsLen = test.FOptsLen
				Convey(fmt.Sprintf("Then MarshalBinary returns %v", test.Bytes), func() {
					b, err := c.MarshalBinary()
					So(err, ShouldBeNil)
					So(b, ShouldResemble, test.Bytes)
				})
			})

			Convey(fmt.Sprintf("Given the slice %v", test.Bytes), func() {
				b := test.Bytes
				Convey(fmt.Sprintf("Then UnmarshalBinary returns a FCtrl with ADR=%v, ADRACKReq=%v, ACK=%v, FPending=%v, FOptsLen=%d", test.ADR, test.ADRACKReq, test.ACK, test.FPending, test.FOptsLen), func() {
					err := c.UnmarshalBinary(b)
					So(err, ShouldBeNil)
					So(c, ShouldResemble, FCtrl{ADR: test.ADR, ADRACKReq: test.ADRACKReq, ACK: test.ACK, FPending: test.FPending, FOptsLen: test.FOptsLen})
				})
			})
		}
	})
}
