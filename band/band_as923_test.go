package band

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAS923Band(t *testing.T) {
	Convey("Given the AS 923 band is selected and 400ms dwell-time is set", t, func() {
		band, err := GetConfig(AS_923, true, lorawan.DwellTime400ms)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     923200000,
				RX2DataRate:      2,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the exepcted value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 14)
		})

		Convey("Then GetPingSlotFrequency returns the expected value", func() {
			freq, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(freq, ShouldEqual, 923400000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(2)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 2)
		})

		Convey("Then RX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(923200000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 923200000)
		})

		Convey("When testing GetRX1DataRateIndex", func() {
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
						dr, err := band.GetRX1DataRateIndex(test.UplinkDR, test.RX1DROffset)
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

		Convey("When testing GetRX1DataRateIndex", func() {
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
						dr, err := band.GetRX1DataRateIndex(test.UplinkDR, test.RX1DROffset)
						So(err, ShouldBeNil)
						So(dr, ShouldEqual, test.ExpectedDR)
					})
				})
			}
		})
	})
}
