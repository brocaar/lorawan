package lorawan

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMHDR(t *testing.T) {
	Convey("Given an empty MHDR", t, func() {
		var mhdr MHDR
		Convey("The MType() = JoinRequest", func() {
			So(mhdr.MType(), ShouldEqual, JoinRequest)
		})
		Convey("The Major() = LoRaWANR1", func() {
			So(mhdr.Major(), ShouldEqual, LoRaWANR1)
		})
	})

	Convey("Given NewMHDR(UnconfirmedDataUp,MajorRFU3)", t, func() {
		mhdr := NewMHDR(UnconfirmedDataUp, MajorRFU3)
		Convey("The MType() = UnconfirmedDataUp", func() {
			So(mhdr.MType(), ShouldEqual, UnconfirmedDataUp)
		})
		Convey("The Major() = MajorRFU3", func() {
			So(mhdr.Major(), ShouldEqual, MajorRFU3)
		})
	})
}
