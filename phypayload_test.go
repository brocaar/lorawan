package lorawan

import (
	"errors"
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
