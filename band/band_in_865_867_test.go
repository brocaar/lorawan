package band

import (
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIN865Band(t *testing.T) {
	Convey("Given the IN 865 band is selected with repeaterCompatible=true", t, func() {
		band, err := GetConfig(IN_865_867, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     866550000,
				RX2DataRate:      2,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 27)
		})

		Convey("Then GetPingSlotFrequency returns the exepected value", func() {
			f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 866550000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(2)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 2)
		})

		Convey("Then RX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(866550000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 866550000)
		})

		Convey("Then the max payload size (N) is 222 for DR4", func() {
			s, err := band.GetMaxPayloadSizeForDataRateIndex(LoRaWAN_1_0_2, RegParamRevB, 4)
			So(err, ShouldBeNil)
			So(s.N, ShouldEqual, 222)
		})
	})

	Convey("Given the IN 865 band is selected with repeaterCompatible=false", t, func() {
		band, err := GetConfig(IN_865_867, false, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then the max payload size (N) is 242 for DR4", func() {
			s, err := band.GetMaxPayloadSizeForDataRateIndex(LoRaWAN_1_0_2, RegParamRevB, 4)
			So(err, ShouldBeNil)
			So(s.N, ShouldEqual, 242)
		})
	})
}
