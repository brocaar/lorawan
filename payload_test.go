package lorawan

import (
	"database/sql/driver"
	"errors"
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
			nonce = DevNonce{1, 2}

			Convey("Then MarshalText returns the expected value", func() {
				b, err := nonce.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "0102")
			})

			Convey("Then Value returns the exepcted value", func() {
				v, err := nonce.Value()
				So(err, ShouldBeNil)
				So(v, ShouldResemble, driver.Value(nonce[:]))
			})
		})

		Convey("Then UnmarshalText sets the nonde correctly", func() {
			So(nonce.UnmarshalText([]byte("0102")), ShouldBeNil)
			So(nonce[:], ShouldResemble, []byte{1, 2})
		})

		Convey("Then Scan sets the nonce correctly", func() {
			So(nonce.Scan([]byte{1, 2}), ShouldBeNil)
			So(nonce[:], ShouldResemble, []byte{1, 2})
		})
	})
}

func TestAppNonce(t *testing.T) {
	Convey("Given an empty AppNonce", t, func() {
		var nonce AppNonce

		Convey("When setting the app-nonce", func() {
			nonce = AppNonce{1, 2, 3}

			Convey("Then MarshalText returns the expected value", func() {
				b, err := nonce.MarshalText()
				So(err, ShouldBeNil)
				So(string(b), ShouldEqual, "010203")
			})

			Convey("Then Value returns the exepcted value", func() {
				v, err := nonce.Value()
				So(err, ShouldBeNil)
				So(v, ShouldResemble, driver.Value(nonce[:]))
			})
		})

		Convey("Then UnmarshalText sets the nonde correctly", func() {
			So(nonce.UnmarshalText([]byte("010203")), ShouldBeNil)
			So(nonce[:], ShouldResemble, []byte{1, 2, 3})
		})

		Convey("Then Scan sets the nonce correctly", func() {
			So(nonce.Scan([]byte{1, 2, 3}), ShouldBeNil)
			So(nonce[:], ShouldResemble, []byte{1, 2, 3})
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
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldResemble, errors.New("lorawan: 18 bytes of data are expected"))
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}", func() {
			b := []byte{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3}
			Convey("Then UnmarshalBinary returns a JoinRequestPayload with AppEUI=[8]byte{1, 1, 1, 1, 1, 1, 1, 1}, DevEUI=[8]byte{2, 2, 2, 2, 2, 2, 2, 2} and DevNonce=[2]byte{3, 3}", func() {
				err := p.UnmarshalBinary(true, b)
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

func TestCFList(t *testing.T) {
	// marshal / unmarshal is already covered by JoinAccept test-case
	Convey("Given an empty CFList", t, func() {
		var l CFList

		Convey("Then each frequency must be a multiple of 100", func() {
			l[0] = 99
			_, err := l.MarshalBinary()
			So(err, ShouldResemble, errors.New("lorawan: frequency must be a multiple of 100"))
			l[0] = 100
			_, err = l.MarshalBinary()
			So(err, ShouldBeNil)
		})

		Convey("Then the frequency values must not exceed 2^24-1 * 100", func() {
			l[0] = 1677721500
			_, err := l.MarshalBinary()
			So(err, ShouldBeNil)
			l[0] = 1677721600
			_, err = l.MarshalBinary()
			So(err, ShouldResemble, errors.New("lorawan: max value of frequency is 2^24-1"))
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

		Convey("Given AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9", func() {
			p.AppNonce = [3]byte{1, 1, 1}
			p.NetID = [3]byte{2, 2, 2}
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

		Convey("Given AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9, CFList=867.1, 867.3, 867.5, 867.7, 867.9", func() {
			p.AppNonce = [3]byte{1, 1, 1}
			p.NetID = [3]byte{2, 2, 2}
			p.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			p.DLSettings.RX2DataRate = 7
			p.DLSettings.RX1DROffset = 6
			p.RXDelay = 9
			p.CFList = &CFList{
				867100000,
				867300000,
				867500000,
				867700000,
				867900000,
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
			Convey("Then UnmarshalBinary returns a JoinAcceptPayload with AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				So(p.AppNonce, ShouldResemble, AppNonce{1, 1, 1})
				So(p.NetID, ShouldResemble, NetID{2, 2, 2})
				So(p.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				So(p.DLSettings, ShouldResemble, DLSettings{RX2DataRate: 7, RX1DROffset: 6})
				So(p.RXDelay, ShouldEqual, 9)
			})
		})

		Convey("Given the slice []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9,24,79,132,232,86,132,184,94,132,136,102,132,88,110,132,0}", func() {
			b := []byte{1, 1, 1, 2, 2, 2, 4, 3, 2, 1, 103, 9, 24, 79, 132, 232, 86, 132, 184, 94, 132, 136, 102, 132, 88, 110, 132, 0}
			Convey("Then UnmarshalBinary returns a JoinAcceptPayload with AppNonce=[3]byte{1, 1, 1}, NetID=[3]byte{2, 2, 2}, DevAddr=DevAddr([4]byte{1, 2, 3, 4}), DLSettings=(RX2DataRate=7, RX1DROffset=6), RXDelay=9, CFlist= 867.1,867.3,867.5,867.7,867.9", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				So(p.AppNonce, ShouldResemble, AppNonce{1, 1, 1})
				So(p.NetID, ShouldResemble, NetID{2, 2, 2})
				So(p.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				So(p.DLSettings, ShouldResemble, DLSettings{RX2DataRate: 7, RX1DROffset: 6})
				So(p.RXDelay, ShouldEqual, 9)
				So(p.CFList, ShouldNotBeNil)
				So(p.CFList, ShouldResemble, &CFList{
					867100000,
					867300000,
					867500000,
					867700000,
					867900000,
				})
			})
		})
	})
}
