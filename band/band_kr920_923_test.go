package band

import (
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKR920Band(t *testing.T) {
	Convey("Given the KR 920 band is selected", t, func() {
		band, err := GetConfig(KR_920_923, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetRX1Channel returns the uplink channel", func() {
			for i := range band.UplinkChannels {
				c := band.GetRX1Channel(i)
				So(c, ShouldEqual, i)
			}
		})

		Convey("Then GetRX1Frequency returns the uplink frequency", func() {
			for _, c := range band.UplinkChannels {
				f, err := band.GetRX1Frequency(c.Frequency)
				So(err, ShouldBeNil)
				So(f, ShouldEqual, c.Frequency)
			}
		})
	})
}
