package band

import (
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCN470Band(t *testing.T) {
	Convey("Given the CN 470-510 band is selected", t, func() {
		band, err := GetConfig(CN_470_510, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("When testing the uplink channels", func() {
			testTable := []struct {
				Channel   int
				Frequency int
				DataRates []int
			}{
				{Channel: 0, Frequency: 470300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
				{Channel: 95, Frequency: 489300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then channel %d must have frequency %d and data rates %v", test.Channel, test.Frequency, test.DataRates), func() {
					So(band.UplinkChannels[test.Channel].Frequency, ShouldEqual, test.Frequency)
					So(band.UplinkChannels[test.Channel].DataRates, ShouldResemble, test.DataRates)
				})
			}
		})

		Convey("When testing the downlink channels", func() {
			testTable := []struct {
				Frequency    int
				Channel      int
				ExpFrequency int
			}{
				{Frequency: 470300000, Channel: 0, ExpFrequency: 500300000},
				{Frequency: 489300000, Channel: 95, ExpFrequency: 509700000},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then frequency: %d must return frequency: %d", test.Frequency, test.ExpFrequency), func() {
					txChan, err := band.GetChannel(test.Frequency, nil)
					So(err, ShouldBeNil)
					So(txChan, ShouldEqual, test.Channel)

					freq, err := band.GetRX1Frequency(test.Frequency)
					So(err, ShouldBeNil)
					So(freq, ShouldEqual, test.ExpFrequency)
				})
			}
		})

		Convey("When iterating over all data rates", func() {
			notImplemented := DataRate{}
			for i, d := range band.DataRates {
				if d == notImplemented {
					continue
				}

				Convey(fmt.Sprintf("Then %v should be DR%d (test %d)", d, i, i), func() {
					dr, err := band.GetDataRate(d)
					So(err, ShouldBeNil)
					So(dr, ShouldEqual, i)
				})
			}
		})
	})
}
