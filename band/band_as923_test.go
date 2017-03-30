package band

import (
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAS923Band(t *testing.T) {
	Convey("Given the AS 923 band is selected and 400ms dwell-time is set", t, func() {
		band, err := GetConfig(AS_923, true, lorawan.DwellTime400ms)
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

		Convey("When testing GetRX1DataRate", func() {
			tests := []struct {
				UplinkDR    int
				RX1DROffset int
				ExpectedDR  int
			}{
				{5, 0, 5},
				{5, 1, 4},
				{5, 2, 3},
				{5, 3, 2},
				{5, 4, 2},
				{5, 5, 2},
				{5, 6, 5},
				{5, 7, 5},
				{2, 6, 3},
				{2, 7, 4},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("When UplinkDR: %d and RX1DROffset: %d [%d]", test.UplinkDR, test.RX1DROffset, i), func() {
					Convey(fmt.Sprintf("Then DownlinkDR: %d", test.ExpectedDR), func() {
						dr, err := band.GetRX1DataRate(test.UplinkDR, test.RX1DROffset)
						So(err, ShouldBeNil)
						So(dr, ShouldEqual, test.ExpectedDR)
					})
				})
			}
		})
	})

	Convey("Given the AS 923 band is selected and no dwell-time is set", t, func() {
		band, err := GetConfig(AS_923, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("When testing GetRX1DataRate", func() {
			tests := []struct {
				UplinkDR    int
				RX1DROffset int
				ExpectedDR  int
			}{
				{5, 0, 5},
				{5, 1, 4},
				{5, 2, 3},
				{5, 3, 2},
				{5, 4, 1},
				{5, 5, 0},
				{5, 6, 5},
				{5, 7, 5},
				{2, 6, 3},
				{2, 7, 4},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("When UplinkDR: %d and RX1DROffset: %d [%d]", test.UplinkDR, test.RX1DROffset, i), func() {
					Convey(fmt.Sprintf("Then DownlinkDR: %d", test.ExpectedDR), func() {
						dr, err := band.GetRX1DataRate(test.UplinkDR, test.RX1DROffset)
						So(err, ShouldBeNil)
						So(dr, ShouldEqual, test.ExpectedDR)
					})
				})
			}
		})
	})
}
