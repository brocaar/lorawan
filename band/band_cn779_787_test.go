package band

import (
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCN779Band(t *testing.T) {
	Convey("Given the CN 779 band is selected with repeaterCompatible=true", t, func() {
		band, err := GetConfig(CN_779_787, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     786000000,
				RX2DataRate:      0,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 10)
		})

		Convey("Then GetPingSlotFrequency returns the expected value", func() {
			f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 785000000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(3)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 3)
		})

		Convey("Then GetRX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(779500000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 779500000)
		})

		Convey("Then the max payload size (N) is 222 for DR4", func() {
			s, err := band.GetMaxPayloadSizeForDataRateIndex(LoRaWAN_1_0_2, RegParamRevB, 4)
			So(err, ShouldBeNil)
			So(s.N, ShouldEqual, 222)
		})
	})

	Convey("Given the CN 779 band is selected with repeaterCompatible=false", t, func() {
		band, err := GetConfig(CN_779_787, false, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then the max payload size (N) is 242 for DR4", func() {
			s, err := band.GetMaxPayloadSizeForDataRateIndex(LoRaWAN_1_0_2, RegParamRevB, 4)
			So(err, ShouldBeNil)
			So(s.N, ShouldEqual, 242)
		})
	})
}
