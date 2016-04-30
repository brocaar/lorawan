package band

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAU915Band(t *testing.T) {
	Convey("Given the AU 915-928 band is selected", t, func() {
		band, err := GetConfig(AU_915_928)
		So(err, ShouldBeNil)

		Convey("When testing the uplink channels", func() {
			testTable := []struct {
				Channel   int
				Frequency int
				DataRates []int
			}{
				{Channel: 0, Frequency: 915200000, DataRates: []int{0, 1, 2, 3}},
				{Channel: 63, Frequency: 927800000, DataRates: []int{0, 1, 2, 3}},
				{Channel: 64, Frequency: 915900000, DataRates: []int{4}},
				{Channel: 71, Frequency: 927100000, DataRates: []int{4}},
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
				Err          error
			}{
				{Frequency: 915900000, DataRate: 4, Channel: 64, ExpFrequency: 923300000},
				{Frequency: 915900000, DataRate: 3, Channel: 0, Err: errors.New("lorawan/band: unknown channel for frequency: 915900000 and data-rate: 3")},
				{Frequency: 915200000, DataRate: 3, Channel: 0, ExpFrequency: 923300000},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then frequency: %d and data rate: %d must return frequency: %d or error: %v", test.Frequency, test.DataRate, test.ExpFrequency, test.Err), func() {
					txChan, err := band.GetChannel(test.Frequency, test.DataRate)

					if test.Err != nil {
						So(err, ShouldResemble, test.Err)
					} else {
						So(err, ShouldBeNil)
						So(txChan, ShouldEqual, test.Channel)
						rx1Chan := band.GetRX1Channel(txChan)
						So(band.DownlinkChannels[rx1Chan].Frequency, ShouldEqual, test.ExpFrequency)
					}
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
