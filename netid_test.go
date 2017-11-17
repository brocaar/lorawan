package lorawan

import (
	"database/sql/driver"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetID(t *testing.T) {
	Convey("Given an empty NetID", t, func() {
		var netID NetID

		Convey("When the value is [3]{1, 2, 219}", func() {
			netID = [3]byte{1, 2, 219}

			Convey("Then MarshalText returns 0102db", func() {
				b, err := netID.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "0102db")
			})

			Convey("Then NwkID returns byte(91)", func() {
				So(netID.NwkID(), ShouldEqual, byte(91))
			})

			Convey("Then Value returns the expected value", func() {
				v, err := netID.Value()
				So(err, ShouldBeNil)
				So(v, ShouldResemble, driver.Value(netID[:]))
			})
		})

		Convey("Given the string 0102db", func() {
			str := "0102db"
			Convey("Then UnmarshalText returns NetID{1, 2, 219}", func() {
				err := netID.UnmarshalText([]byte(str))
				So(err, ShouldBeNil)
				So(netID, ShouldEqual, NetID{1, 2, 219})
			})
		})

		Convey("Given a byteslice", func() {
			b := []byte{1, 2, 3}
			Convey("Then Scan scans the value correctly", func() {
				So(netID.Scan(b), ShouldBeNil)
				So(netID[:], ShouldResemble, b)
			})
		})
	})
}
