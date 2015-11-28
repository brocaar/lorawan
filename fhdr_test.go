package lorawan

import (
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
		var fc FCtrl
		Convey("ADR, ADRACKReq, ACK and FPending should be false", func() {
			So(fc.ADR(), ShouldBeFalse)
			So(fc.ADRACKReq(), ShouldBeFalse)
			So(fc.ACK(), ShouldBeFalse)
			So(fc.FPending(), ShouldBeFalse)
		})
		Convey("FOptsLen = 0", func() {
			So(fc.FOptsLen(), ShouldEqual, 0)
		})
	})

	Convey("Given I use NewFCtrl to create a new FCtrl", t, func() {
		Convey("An error should be returned when fOptsLen > 15 should", func() {
			_, err := NewFCtrl(false, false, false, false, 16)
			So(err, ShouldNotBeNil)
		})
		Convey("ADR() == true when adr is set", func() {
			fc, err := NewFCtrl(true, false, false, false, 0)
			So(err, ShouldBeNil)
			So(fc.ADR(), ShouldBeTrue)
			So(fc.ADRACKReq(), ShouldBeFalse)
			So(fc.ACK(), ShouldBeFalse)
			So(fc.FPending(), ShouldBeFalse)
		})
		Convey("ADRACKReq() == true when adrAckReq is set", func() {
			fc, err := NewFCtrl(false, true, false, false, 0)
			So(err, ShouldBeNil)
			So(fc.ADRACKReq(), ShouldBeTrue)
			So(fc.ADR(), ShouldBeFalse)
			So(fc.ACK(), ShouldBeFalse)
			So(fc.FPending(), ShouldBeFalse)

		})
		Convey("ACK() == true when ack is set", func() {
			fc, err := NewFCtrl(false, false, true, false, 0)
			So(err, ShouldBeNil)
			So(fc.ACK(), ShouldBeTrue)
			So(fc.ADR(), ShouldBeFalse)
			So(fc.ADRACKReq(), ShouldBeFalse)
			So(fc.FPending(), ShouldBeFalse)
		})
		Convey("FPending() == true when fPending is set", func() {
			fc, err := NewFCtrl(false, false, false, true, 0)
			So(err, ShouldBeNil)
			So(fc.FPending(), ShouldBeTrue)
			So(fc.ADR(), ShouldBeFalse)
			So(fc.ADRACKReq(), ShouldBeFalse)
			So(fc.ACK(), ShouldBeFalse)

		})
		Convey("FOptsLen() == 11, when fOptsLen is set to 11", func() {
			fc, err := NewFCtrl(false, false, false, false, 11)
			So(err, ShouldBeNil)
			So(fc.ADR(), ShouldBeFalse)
			So(fc.ADRACKReq(), ShouldBeFalse)
			So(fc.ACK(), ShouldBeFalse)
			So(fc.FPending(), ShouldBeFalse)
			So(fc.FOptsLen(), ShouldEqual, 11)
		})
	})
}
