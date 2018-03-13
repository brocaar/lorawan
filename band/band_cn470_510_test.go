package band

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCN470Band(t *testing.T) {
	Convey("Given the CN 470-510 band is selected", t, func() {
		band, err := GetConfig(CN_470_510, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     505300000,
				RX2DataRate:      0,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 14)
		})

		Convey("Then GetPingSlotFrequency returns the expected value", func() {
			tests := []struct {
				DevAddr           lorawan.DevAddr
				BeaconTime        string
				ExpectedFrequency int
			}{
				{
					DevAddr:           lorawan.DevAddr{3, 20, 207, 54},
					BeaconTime:        "334382h51m44s",
					ExpectedFrequency: 501100000,
				},
			}

			for _, test := range tests {
				bt, err := time.ParseDuration(test.BeaconTime)
				So(err, ShouldBeNil)
				freq, err := band.GetPingSlotFrequency(test.DevAddr, bt)
				So(err, ShouldBeNil)
				So(freq, ShouldEqual, test.ExpectedFrequency)
			}
		})

		Convey("When testing the uplink channels", func() {
			testTable := []struct {
				Channel   int
				Frequency int
				MinDR     int
				MaxDR     int
			}{
				{Channel: 0, Frequency: 470300000, MinDR: 0, MaxDR: 5},
				{Channel: 95, Frequency: 489300000, MinDR: 0, MaxDR: 5},
			}

			for _, test := range testTable {
				Convey(fmt.Sprintf("Then channel %d must have frequency %d and min/max data-rates %d/%d", test.Channel, test.Frequency, test.MinDR, test.MaxDR), func() {
					c, err := band.GetUplinkChannel(test.Channel)
					So(err, ShouldBeNil)
					So(c.Frequency, ShouldEqual, test.Frequency)
					So(c.MinDR, ShouldResemble, test.MinDR)
					So(c.MaxDR, ShouldResemble, test.MaxDR)
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
					txChan, err := band.GetUplinkChannelIndex(test.Frequency, true)
					So(err, ShouldBeNil)
					So(txChan, ShouldEqual, test.Channel)

					freq, err := band.GetRX1FrequencyForUplinkFrequency(test.Frequency)
					So(err, ShouldBeNil)
					So(freq, ShouldEqual, test.ExpFrequency)
				})
			}
		})

		Convey("When testing LinkADRReqPayload functions", func() {
			allChannels := band.GetUplinkChannelIndices()
			var filteredChannels []int

			for i := 0; i < 96; i++ {
				if i == 6 || i == 38 || i == 45 {
					continue
				}
				filteredChannels = append(filteredChannels, i)
			}

			tests := []struct {
				Name                       string
				NodeChannels               []int
				DisableChannels            []int
				ExpectedUplinkChannels     []int
				ExpectedLinkADRReqPayloads []lorawan.LinkADRReqPayload
			}{
				{
					Name:                   "all channels active",
					NodeChannels:           band.GetEnabledUplinkChannelIndices(),
					ExpectedUplinkChannels: allChannels,
				},
				{
					Name:                   "channel 6, 38 and 45 disabled",
					NodeChannels:           band.GetEnabledUplinkChannelIndices(),
					DisableChannels:        []int{6, 38, 45},
					ExpectedUplinkChannels: filteredChannels,
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, false, true, true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 2},
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("testing %s [%d]", test.Name, i), func() {
					for _, c := range test.DisableChannels {
						So(band.DisableUplinkChannelIndex(c), ShouldBeNil)
					}
					pls := band.GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(test.NodeChannels)
					So(pls, ShouldResemble, test.ExpectedLinkADRReqPayloads)

					chans, err := band.GetEnabledUplinkChannelIndicesForLinkADRReqPayloads(test.NodeChannels, pls)
					So(err, ShouldBeNil)
					So(chans, ShouldResemble, test.ExpectedUplinkChannels)
				})
			}
		})
	})
}
