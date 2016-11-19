package band

import (
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUS902Band(t *testing.T) {
	Convey("Given the US 902-928 band is selected", t, func() {
		band, err := GetConfig(US_902_928, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("When testing the uplink channels", func() {
			testTable := []struct {
				Channel   int
				Frequency int
				DataRates []int
			}{
				{Channel: 0, Frequency: 902300000, DataRates: []int{0, 1, 2, 3}},
				{Channel: 63, Frequency: 914900000, DataRates: []int{0, 1, 2, 3}},
				{Channel: 64, Frequency: 903000000, DataRates: []int{4}},
				{Channel: 71, Frequency: 914200000, DataRates: []int{4}},
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
				DataRate     int
				Channel      int
				ExpFrequency int
			}{
				{Frequency: 914900000, DataRate: 3, Channel: 63, ExpFrequency: 927500000},
				{Frequency: 903000000, DataRate: 4, Channel: 64, ExpFrequency: 923300000},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then frequency: %d must return frequency: %d", test.Frequency, test.ExpFrequency), func() {
					txChan, err := band.GetChannel(test.Frequency, nil)
					So(err, ShouldBeNil)
					So(txChan, ShouldEqual, test.Channel)
					rx1Chan := band.GetRX1Channel(txChan)
					freq, err := band.GetDownlinkFrequency(rx1Chan, nil)
					So(err, ShouldBeNil)
					So(freq, ShouldEqual, test.ExpFrequency)

					freq, err = band.GetRX1Frequency(test.Frequency)
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

				expected := i
				if i == 12 {
					expected = 4
				}

				Convey(fmt.Sprintf("Then %v should be DR%d (test %d)", d, expected, i), func() {
					dr, err := band.GetDataRate(d)
					So(err, ShouldBeNil)
					So(dr, ShouldEqual, expected)
				})
			}
		})
	})
}
