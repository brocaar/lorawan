package band

import (
	"errors"
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAU915Band(t *testing.T) {
	Convey("Given the AU 915-928 band is selected", t, func() {
		band, err := GetConfig(AU_915_928, true, lorawan.DwellTimeNoLimit)
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
			}{
				{Frequency: 915900000, DataRate: 4, Channel: 64, ExpFrequency: 923300000},
				{Frequency: 915200000, DataRate: 3, Channel: 0, ExpFrequency: 923300000},
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

		Convey("When testing GetRX1DataRateForOffset", func() {
			testTable := []struct {
				DR       int
				DROffset int
				RX1DR    int
				Error    error
			}{
				{0, 0, 10, nil},
				{0, 1, 9, nil},
				{0, 4, 0, errors.New("lorawan/band: invalid data-rate offset")},
				{4, 0, 13, nil},
				{5, 0, 0, errors.New("lorawan/band: invalid data-rate")},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Given DR %d, DR offset %d", test.DR, test.DROffset), func() {
					dr, err := band.GetRX1DataRate(test.DR, test.DROffset)
					Convey(fmt.Sprintf("Then RX1DR=%d, error=%s", dr, err), func() {
						So(dr, ShouldEqual, test.RX1DR)
						So(err, ShouldResemble, test.Error)
					})
				})
			}
		})
	})
}
