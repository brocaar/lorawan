package lorawan

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
				err := p.UnmarshalBinary(b)
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

		Convey("Given AppEUI=[8]byte{1, 1, 1, 1, 1, 1, 1, 1}, DevEUI=[8]byte{2, 2, 2, 2, 2, 2, 2, 2} and DevNonce=[2]byte{3, 3}", func() {
			p.AppEUI = [8]byte{1, 1, 1, 1, 1, 1, 1, 1}
			p.DevEUI = [8]byte{2, 2, 2, 2, 2, 2, 2, 2}
			p.DevNonce = [2]byte{3, 3}
			Convey("Then MarshalBinary returns []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3})
			})
		})

		Convey("Given a slice of bytes with an invalid size", func() {
			b := make([]byte, 17)
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldResemble, errors.New("lorawan: 18 bytes of data are expected"))
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}", func() {
			b := []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}
			Convey("Then UnmarshalBinary returns a JoinRequestPayload with AppEUI=[8]byte{1, 1, 1, 1, 1, 1, 1, 1}, DevEUI=[8]byte{2, 2, 2, 2, 2, 2, 2, 2} and DevNonce=[2]byte{3, 3}", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, JoinRequestPayload{
					AppEUI:   [8]byte{1, 1, 1, 1, 1, 1, 1, 1},
					DevEUI:   [8]byte{2, 2, 2, 2, 2, 2, 2, 2},
					DevNonce: [2]byte{3, 3},
				})
			})
		})
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

		Convey("Given AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DRoffset=6), RXDelay=9", func() {
			p.AppNonce = [3]byte{1, 1, 1}
			p.NetID = [3]byte{2, 2, 2}
			p.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			p.DLSettings.RX2DataRate = 7
			p.DLSettings.RX1DRoffset = 6
			p.RXDelay = 9

			Convey("Then MarshalBinary returns []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9})
			})
		})

		Convey("Given a slice of bytes with an invalid size", func() {
			b := make([]byte, 11)
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldResemble, errors.New("lorawan: 12 bytes of data are expected"))
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}", func() {
			b := []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9}
			Convey("Then UnmarshalBinary returns a JoinAcceptPayload with AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DRoffset=6), RXDelay=9", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)

				So(p.AppNonce, ShouldResemble, [3]byte{1, 1, 1})
				So(p.NetID, ShouldResemble, [3]byte{2, 2, 2})
				So(p.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				So(p.DLSettings, ShouldResemble, DLsettings{RX2DataRate: 7, RX1DRoffset: 6})
				So(p.RXDelay, ShouldEqual, 9)
			})
		})
	})
}
