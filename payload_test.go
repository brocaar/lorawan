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

		Convey("Given AppEUI=1, DevEUI=2, DevNonce=3", func() {
			p.AppEUI = 1
			p.DevEUI = 2
			p.DevNonce = 3
			Convey("Then MarshalBinary returns []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0})
			})
		})

		Convey("Given the slice []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}", func() {
			b := make([]byte, 17)
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldResemble, errors.New("lorawan: 18 bytes of data are expected"))
			})
		})

		Convey("Given the slice []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0}", func() {
			b := []byte{1, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 3, 0}
			Convey("Then UnmarshalBinary returns a JoinRequestPayload with AppEUI=1, DevEUI=2, DevNonce=3", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(p, ShouldResemble, JoinRequestPayload{AppEUI: 1, DevEUI: 2, DevNonce: 3})
			})
		})
	})
}
