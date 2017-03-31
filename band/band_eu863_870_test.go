package band

import (
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEU863Band(t *testing.T) {
	Convey("Given the EU 863-870 band is selected", t, func() {
		band, err := GetConfig(EU_863_870, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetRX1Channel returns the uplink channel", func() {
			for i := 0; i < 3; i++ {
				rx1Chan := band.GetRX1Channel(i)
				So(rx1Chan, ShouldEqual, i)
			}
		})

		Convey("Then GetRX1Frequency returns the uplink frequency", func() {
			for _, f := range []int{868100000, 868200000, 868300000} {
				freq, err := band.GetRX1Frequency(f)
				So(err, ShouldBeNil)
				So(freq, ShouldEqual, f)
			}
		})

		Convey("Given five extra channels", func() {
			chans := []int{
				867100000,
				867300000,
				867500000,
				867700000,
				867900000,
			}

			for _, c := range chans {
				band.AddChannel(c)
			}

			Convey("Then GetChannel takes the extra channels into consideration", func() {
				tests := []int{
					868100000,
					868300000,
					868500000,
					867100000,
					867300000,
					867500000,
					867700000,
					867900000,
				}

				for expChannel, expFreq := range tests {
					channel, err := band.GetUplinkChannelNumber(expFreq)
					So(err, ShouldBeNil)
					So(channel, ShouldEqual, expChannel)
				}
			})

			Convey("Then GetCFList returns the expected CFList", func() {
				cFList := band.GetCFList()
				So(cFList, ShouldNotBeNil)
				So(*cFList, ShouldEqual, lorawan.CFList{
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
