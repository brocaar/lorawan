package lorawan

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetID(t *testing.T) {
	Convey("Given a set of tests", t, func() {
		tests := []struct {
			Name   string
			NetID  NetID
			Type   int
			ID     []byte
			Bytes  []byte
			String string
		}{
			{
				Name:   "NetID type 0",
				NetID:  NetID{0, 0, 109},
				Type:   0,
				ID:     []byte{45},
				Bytes:  []byte{109, 0, 0},
				String: "00006d",
			},
			{
				Name:   "NetID type 1",
				NetID:  NetID{32, 0, 109},
				Type:   1,
				ID:     []byte{45},
				Bytes:  []byte{109, 0, 32},
				String: "20006d",
			},
			{
				Name:   "NetID type 2",
				NetID:  NetID{64, 3, 109},
				Type:   2,
				ID:     []byte{1, 109},
				Bytes:  []byte{109, 3, 64},
				String: "40036d",
			},
			{
				Name:   "NetID type 3",
				NetID:  NetID{118, 219, 109},
				Type:   3,
				ID:     []byte{22, 219, 109},
				Bytes:  []byte{109, 219, 118},
				String: "76db6d",
			},
			{
				Name:   "NetID type 4",
				NetID:  NetID{150, 219, 109},
				Type:   4,
				ID:     []byte{22, 219, 109},
				Bytes:  []byte{109, 219, 150},
				String: "96db6d",
			},
			{
				Name:   "NetID type 5",
				NetID:  NetID{182, 219, 109},
				Type:   5,
				ID:     []byte{22, 219, 109},
				Bytes:  []byte{109, 219, 182},
				String: "b6db6d",
			},
			{
				Name:   "NetID type 6",
				NetID:  NetID{214, 219, 109},
				Type:   6,
				ID:     []byte{22, 219, 109},
				Bytes:  []byte{109, 219, 214},
				String: "d6db6d",
			},
			{
				Name:   "NetID type 7",
				NetID:  NetID{246, 219, 109},
				Type:   7,
				ID:     []byte{22, 219, 109},
				Bytes:  []byte{109, 219, 246},
				String: "f6db6d",
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Testing: %s [%d]", test.Name, i), func() {
				So(test.NetID.Type(), ShouldEqual, test.Type)
				So(test.NetID.ID(), ShouldResemble, test.ID)

				b, err := test.NetID.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, test.Bytes)

				var netID NetID
				So(netID.UnmarshalBinary(test.Bytes), ShouldBeNil)
				So(netID, ShouldEqual, test.NetID)

				b, err = test.NetID.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, test.String)

				So(netID.UnmarshalText([]byte(test.String)), ShouldBeNil)
				So(netID, ShouldEqual, test.NetID)
			})
		}
	})
}
