package band

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRU864Band(t *testing.T) {
	Convey("Given the RU 864-869 band is selected", t, func() {
		band, err := GetConfig(RU_864_870, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     869100000,
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
			f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 868900000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(3)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 3)
		})

		Convey("Then GetRX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(868900000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 868900000)
		})

		Convey("Given five extra channels", func() {
			chans := []int{
				864100000,
				864300000,
				864500000,
				864700000,
				864900000,
			}

			for _, c := range chans {
				band.AddChannel(c, 0, 5)
			}

			Convey("Then these are returned as custom channels", func() {
				So(band.GetCustomUplinkChannelIndices(), ShouldResemble, []int{2, 3, 4, 5, 6})
			})

			Convey("When testing the LinkADRReqPayload functions", func() {
				tests := []struct {
					Name                       string
					NodeChannels               []int
					DisabledChannels           []int
					ExpectedUplinkChannels     []int
					ExpectedLinkADRReqPayloads []lorawan.LinkADRReqPayload
				}{
					{
						Name:                   "no active node channels",
						NodeChannels:           []int{},
						ExpectedUplinkChannels: []int{0, 1},
						ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
							{
								ChMask: lorawan.ChMask{true, true},
							},
						},
						// we only activate the base channels
					},
					{
						Name:                   "base channels are active",
						NodeChannels:           []int{0, 1},
						ExpectedUplinkChannels: []int{0, 1},
						// we do not activate the CFList channels as we don't
						// now if the node knows about these frequencies
					},
					{
						Name:                   "base channels + two CFList channels are active",
						NodeChannels:           []int{0, 1, 2, 3},
						ExpectedUplinkChannels: []int{0, 1, 2, 3},
						// we do not activate the CFList channels as we don't
						// now if the node knows about these frequencies
					},
					{
						Name:                   "base channels + CFList are active",
						NodeChannels:           []int{0, 1, 2, 3, 4, 5, 6},
						ExpectedUplinkChannels: []int{0, 1, 2, 3, 4, 5, 6},
						// nothing to do, network and node are in sync
					},
					{
						Name:                   "base channels + CFList are active on node, but CFList channels are disabled on the network",
						NodeChannels:           []int{0, 1, 2, 3, 4, 5, 6},
						DisabledChannels:       []int{2, 3, 4, 5, 6},
						ExpectedUplinkChannels: []int{0, 1},
						ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
							{
								ChMask: lorawan.ChMask{true, true},
							},
						},
						// we disable the CFList channels as they became inactive
					},
				}

				for i, test := range tests {
					Convey(fmt.Sprintf("testing %s [%d]", test.Name, i), func() {
						for _, c := range test.DisabledChannels {
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

			Convey("Then GetChannel takes the extra channels into consideration", func() {
				tests := []int{
					868900000,
					869100000,
					864100000,
					864300000,
					864500000,
					864700000,
					864900000,
				}

				for expChannel, expFreq := range tests {
					var defaultChannel bool
					if expChannel < 2 {
						defaultChannel = true
					}
					channel, err := band.GetUplinkChannelIndex(expFreq, defaultChannel)
					So(err, ShouldBeNil)
					So(channel, ShouldEqual, expChannel)
				}
			})

			Convey("Then GetCFList returns the expected CFList", func() {
				cFList := band.GetCFList(LoRaWAN_1_0_2)
				So(cFList, ShouldNotBeNil)
				So(cFList, ShouldResemble, &lorawan.CFList{
					CFListType: lorawan.CFListChannel,
					Payload: &lorawan.CFListChannelPayload{
						Channels: [5]uint32{
							864100000,
							864300000,
							864500000,
							864700000,
							864900000,
						},
					},
				})
			})
		})
	})
}
