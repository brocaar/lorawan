package lorawan

import (
	"bytes"
	"errors"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMACPayload(t *testing.T) {
	Convey("Given an empty MACPayload", t, func() {
		var p MACPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given FPort=1", func() {
			fPort := uint8(1)
			p.FPort = &fPort

			Convey("Given FRMPayload contains a MACCommand", func() {
				p.FRMPayload = []Payload{&MACCommand{CID: LinkCheckReq}}
				Convey("Then MarshalBinary returns an error that FPort must be 0", func() {
					_, err := p.MarshalBinary()
					So(err, ShouldResemble, errors.New("lorawan: a MAC command is only allowed when FPort=0"))
				})
			})
		})

		Convey("Given FPort=nil", func() {

			Convey("Given FRMPayload is not empty", func() {
				p.FRMPayload = []Payload{&DataPayload{Bytes: []byte{1}}}
				Convey("Then MarshalBinary returns an error that FPort must be set", func() {
					_, err := p.MarshalBinary()
					So(err, ShouldResemble, errors.New("lorawan: FPort must be set when FRMPayload is not empty"))
				})
			})
		})

		Convey("Given FPort=0", func() {
			fPort := uint8(0)
			p.FPort = &fPort

			Convey("Given FOpts are set", func() {
				p.FHDR.FOpts = []Payload{&MACCommand{CID: LinkCheckReq}}
				Convey("Then MarshalBinary returns an error that FPort must not be 0", func() {
					_, err := p.MarshalBinary()
					So(err, ShouldResemble, errors.New("lorawan: FPort must not be 0 when FOpts are set"))
				})
			})
		})

		Convey("Given FHDR(DevAddr=[4]{1, 2, 3, 4}), FPort=1, FRMPayload=[]Payload{DataPayload(Bytes=[]byte{5, 6, 7})}", func() {
			p.FHDR.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			fPort := uint8(1)
			p.FPort = &fPort
			p.FRMPayload = []Payload{&DataPayload{[]byte{5, 6, 7}}}

			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1, 0, 0, 0, 1, 5, 6, 7}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1, 0, 0, 0, 1, 5, 6, 7})
			})
		})

		Convey("Given uplink=true, FHDR(DevAddr=[4]{1, 2, 3, 4}), FPort=0, FRMPayload=[]Payload{MACCommand{CID: DevStatusAns, Payload: DevStatusAnsPayload(Battery=10, Margin=20)}}", func() {
			p.FHDR.DevAddr = DevAddr([4]byte{1, 2, 3, 4})
			fPort := uint8(0)
			p.FPort = &fPort
			p.FRMPayload = []Payload{&MACCommand{CID: DevStatusAns, Payload: &DevStatusAnsPayload{Battery: 10, Margin: 20}}}

			Convey("Then MarshalBinary returns []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20})
			})
		})

		Convey("Given the slice []byte{4, 3, 2, 1, 0, 0}", func() {
			b := []byte{4, 3, 2, 1, 0, 0}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldResemble, errors.New("lorawan: at least 7 bytes needed to decode FHDR"))
			})
		})

		Convey("Given the slice []byte{4, 3, 2, 1, 3, 0, 0, 0, 0}", func() {
			b := []byte{4, 3, 2, 1, 3, 0, 0, 0, 0}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldResemble, errors.New("lorawan: not enough bytes to decode FHDR"))
			})
		})

		Convey("Given uplink=true and slice []byte{4, 3, 2, 1, 1, 0, 0, 2}", func() {
			b := []byte{4, 3, 2, 1, 1, 0, 0, 2}
			Convey("Then UnmarshalBinary returns no error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldBeNil)
			})
		})

		Convey("Given uplink=true and slice []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10}", func() {
			b := []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldBeNil)

				// normally the mac commands are unmarshaled after decryption
				_, err = decodeDataPayloadToMACCommands(true, p.FRMPayload)
				So(err, ShouldResemble, errors.New("lorawan: not enough remaining bytes"))
			})
		})

		Convey("Given uplink=true and slice []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20}", func() {
			b := []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20}

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldBeNil)

				Convey("Then FHDR(DevAddr=[4]byte{1, 2, 3, 4})", func() {
					So(p.FHDR.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})
				Convey("Then FPort=0", func() {
					So(p.FPort, ShouldNotBeNil)
					So(*p.FPort, ShouldEqual, 0)
				})
				Convey("Then FRMPayload=[]Payload{MACCommand{CID: DevStatusAns, Payload: DevStatusAnsPayload(Battery=10, Margin=20)}}", func() {
					// mac commands are normally unmarshaled when decrypting
					var err error
					p.FRMPayload, err = decodeDataPayloadToMACCommands(true, p.FRMPayload)
					So(err, ShouldBeNil)

					So(p.FRMPayload, ShouldHaveLength, 1)
					mac, ok := p.FRMPayload[0].(*MACCommand)
					So(ok, ShouldBeTrue)
					So(mac.CID, ShouldEqual, DevStatusAns)

					pl, ok := mac.Payload.(*DevStatusAnsPayload)
					So(ok, ShouldBeTrue)
					So(pl.Battery, ShouldEqual, 10)
					So(pl.Margin, ShouldEqual, 20)
				})
			})
		})

		Convey("Given uplink=true and slice []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20, 78, 79} (one known and some unknown data)", func() {
			b := []byte{4, 3, 2, 1, 0, 0, 0, 0, 6, 10, 20, 78, 79}
			var logBytes bytes.Buffer
			log.SetOutput(&logBytes)

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := p.UnmarshalBinary(true, b)
				So(err, ShouldBeNil)

				// mac commands are normally unmarshaled when decrypting
				p.FRMPayload, err = decodeDataPayloadToMACCommands(true, p.FRMPayload)
				So(err, ShouldBeNil)

				Convey("Then FHDR(DevAddr=[4]byte{1, 2, 3, 4})", func() {
					So(p.FHDR.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})
				Convey("Then FPort=0", func() {
					So(p.FPort, ShouldNotBeNil)
					So(*p.FPort, ShouldEqual, 0)
				})
				Convey("Then FRMPayload=[]Payload{MACCommand{CID: DevStatusAns, Payload: DevStatusAnsPayload(Battery=10, Margin=20)}}", func() {

					So(p.FRMPayload, ShouldHaveLength, 3)
					mac, ok := p.FRMPayload[0].(*MACCommand)
					So(ok, ShouldBeTrue)
					So(mac.CID, ShouldEqual, DevStatusAns)

					pl, ok := mac.Payload.(*DevStatusAnsPayload)
					So(ok, ShouldBeTrue)
					So(pl.Battery, ShouldEqual, 10)
					So(pl.Margin, ShouldEqual, 20)

					// unparsable mac-data
					So(p.FRMPayload[1].(*MACCommand).CID, ShouldEqual, 78)
					So(p.FRMPayload[2].(*MACCommand).CID, ShouldEqual, 79)
				})
			})
		})

		Convey("Given the slice []byte{4,3, 2, 1, 0, 0, 0, 1, 6, 10, 20}", func() {
			b := []byte{4, 3, 2, 1, 0, 0, 0, 1, 6, 10, 20}

			Convey("Then UnmarshalBinary does not return an error", func() {
				err := p.UnmarshalBinary(false, b)
				So(err, ShouldBeNil)

				Convey("Then FHDR(DevAddr=[4]byte{1, 2, 3, 4})", func() {
					So(p.FHDR.DevAddr, ShouldEqual, DevAddr([4]byte{1, 2, 3, 4}))
				})
				Convey("Then FPort=1", func() {
					So(p.FPort, ShouldNotBeNil)
					So(*p.FPort, ShouldEqual, 1)
				})
				Convey("Then FRMPayload=[]Payload{DataPayload([]byte{6, 10, 20})}", func() {
					So(p.FRMPayload, ShouldHaveLength, 1)

					pl, ok := p.FRMPayload[0].(*DataPayload)
					So(ok, ShouldBeTrue)
					So(pl.Bytes, ShouldResemble, []byte{6, 10, 20})
				})
			})
		})
	})
}
