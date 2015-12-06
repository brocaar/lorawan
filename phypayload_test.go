package lorawan

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMHDR(t *testing.T) {
	Convey("Given an empty MHDR", t, func() {
		var h MHDR
		Convey("Then MarshalBinary returns []byte{0}", func() {
			b, err := h.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0})
		})

		Convey("Given MType=Proprietary, Major=LoRaWANR1", func() {
			h.MType = Proprietary
			h.Major = LoRaWANR1
			Convey("Then MarshalBinary returns []byte{224}", func() {
				b, err := h.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{224})
			})
		})

		Convey("Given a slice []byte{224}", func() {
			b := []byte{224}
			Convey("Then UnmarshalBinary returns a MHDR with MType=Proprietary, Major=LoRaWANR1", func() {
				err := h.UnmarshalBinary(b)
				So(err, ShouldBeNil)
				So(h, ShouldResemble, MHDR{MType: Proprietary, Major: LoRaWANR1})
			})
		})
	})
}

func TestPHYPayload(t *testing.T) {
	Convey("Given an empty PHYPayload", t, func() {
		var p PHYPayload
		Convey("Then MarshalBinary returns []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}", func() {
			b, err := p.MarshalBinary()
			So(err, ShouldBeNil)
			So(b, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		})

		Convey("Given MHDR(MType=JoinAccept, Major=LoRaWANR1), MACPayload(FHDR(DevAddr=67305985)), MIC=[4]byte{4, 3, 2, 1}", func() {
			p.MHDR.MType = JoinAccept
			p.MHDR.Major = LoRaWANR1
			p.MACPayload.FHDR.DevAddr = DevAddr(67305985)
			p.MIC = [4]byte{4, 3, 2, 1}

			Convey("Given the NwkSKey []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}", func() {
				nwkSKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

				Convey("Then ValidateMIC returns false", func() {
					v, err := p.ValidateMIC(nwkSKey)
					So(err, ShouldBeNil)
					So(v, ShouldBeFalse)

				})

				// todo: check if this mic is valid
				Convey("calculateMIC returns []byte{0xac, 0x80, 0x37, 0x24}", func() {
					mic, err := p.calculateMIC(nwkSKey)
					So(err, ShouldBeNil)
					So(mic, ShouldResemble, []byte{0xac, 0x80, 0x37, 0x24})
				})

				Convey("Given the MIC is []byte{0xac, 0x80, 0x37, 0x24}", func() {
					p.MIC = [4]byte{0xac, 0x80, 0x37, 0x24}

					Convey("Then ValidateMIC returns true", func() {
						v, err := p.ValidateMIC(nwkSKey)
						So(err, ShouldBeNil)
						So(v, ShouldBeTrue)
					})
				})
			})

			Convey("Then MarshalBinary returns []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2, 1}", func() {
				b, err := p.MarshalBinary()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2, 1})
			})
		})

		Convey("Ginve the slice []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2}", func() {
			b := []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2}
			Convey("Then UnmarshalBinary returns an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldResemble, errors.New("lorawan: at least 12 bytes needed to decode PHYPayload"))
			})
		})

		Convey("Given the slice []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2, 1}", func() {
			b := []byte{32, 1, 2, 3, 4, 0, 0, 0, 4, 3, 2, 1}
			Convey("Then UnmarshalBinary does not return an error", func() {
				err := p.UnmarshalBinary(b)
				So(err, ShouldBeNil)

				Convey("Then MHDR=(MType=JoinAccept, Major=LoRaWANR1)", func() {
					So(p.MHDR, ShouldResemble, MHDR{MType: JoinAccept, Major: LoRaWANR1})
				})
				Convey("Then MACPayload(FHDR(DevAddr=67305985))", func() {
					So(p.MACPayload, ShouldResemble, MACPayload{FHDR: FHDR{DevAddr: DevAddr(67305985)}})
				})
				Convey("Then MIC=[4]byte{4, 3, 2, 1}", func() {
					So(p.MIC, ShouldResemble, [4]byte{4, 3, 2, 1})
				})
			})
		})
	})
}

func ExampleNew() {
	uplink := true

	nwkSKey := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := []byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	pl := New(uplink)

	pl.MHDR = MHDR{
		MType: ConfirmedDataUp,
		Major: LoRaWANR1,
	}
	pl.MACPayload = MACPayload{
		FHDR: FHDR{
			DevAddr: DevAddr(67305985),
			FCtrl: FCtrl{
				ADR:       false,
				ADRACKReq: false,
				ACK:       false,
			},
			Fcnt:  0,
			FOpts: []MACCommand{},
		},
		FPort:      10,
		FRMPayload: []Payload{&DataPayload{Bytes: []byte{1, 2, 3, 4}}},
	}

	if err := pl.SetMIC(nwkSKey); err != nil {
		panic(err)
	}
	if err := pl.MACPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	bytes, err := pl.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fmt.Println(bytes)

	// Output:
	// [128 1 2 3 4 0 0 0 10 59 85 197 241 99 239 222 68]
}
