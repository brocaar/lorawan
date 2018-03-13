package band

import (
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestKR920Band(t *testing.T) {
	Convey("Given the KR 920 band is selected", t, func() {
		band, err := GetConfig(KR_920_923, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     921900000,
				RX2DataRate:      0,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 23)
		})

		Convey("Then GetPingSlotFrequency returns the expected value", func() {
			f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 923100000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(2)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 2)
		})

		Convey("Then GetRX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(922100000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 922100000)
		})
	})
}
