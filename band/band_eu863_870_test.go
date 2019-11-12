package band

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEU863Band(t *testing.T) {
	Convey("Given the EU 863-870 band is selected", t, func() {
		band, err := GetConfig(EU_863_870, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     869525000,
				RX2DataRate:      0,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 863.000000 MHz", func() {
			So(band.GetDownlinkTXPower(863000000), ShouldEqual, 14)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 863.000001 MHz", func() {
			So(band.GetDownlinkTXPower(863000001), ShouldEqual, 14)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 869.200000 MHz", func() {
			So(band.GetDownlinkTXPower(869200000), ShouldEqual, 14)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 869.400000 MHz", func() {
			So(band.GetDownlinkTXPower(869400000), ShouldEqual, 27)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 869.400001 MHz", func() {
			So(band.GetDownlinkTXPower(869400001), ShouldEqual, 27)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for 869.650000 MHz", func() {
			So(band.GetDownlinkTXPower(869650000), ShouldEqual, 14)
		})

		Convey("Then GetDownlinkTXPower returns the expected value for any other value (0)", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 14)
		})

		Convey("Then GetPingSlotFrequency returns the expected value", func() {
			f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 869525000)
		})

		Convey("Then GetRX1ChannelIndexForUplinkChannelIndex returns the expected value", func() {
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(3)
			So(err, ShouldBeNil)
			So(c, ShouldEqual, 3)
		})

		Convey("Then GetRX1FrequencyForUplinkFrequency returns the expected value", func() {
			f, err := band.GetRX1FrequencyForUplinkFrequency(868500000)
			So(err, ShouldBeNil)
			So(f, ShouldEqual, 868500000)
		})

		Convey("Given five extra channels", func() {
			chans := []int{
				867100000,
				867300000,
				867500000,
				867700000,
				867900000,
			}

			for _, c := range chans {
				band.AddChannel(c, 0, 5)
			}

			Convey("Then these are returned as custom channels", func() {
				So(band.GetCustomUplinkChannelIndices(), ShouldResemble, []int{3, 4, 5, 6, 7})
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
						ExpectedUplinkChannels: []int{0, 1, 2},
						ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
							{
								ChMask: lorawan.ChMask{true, true, true},
							},
						},
						// we only activate the base channels
					},
					{
						Name:                   "base channels are active",
						NodeChannels:           []int{0, 1, 2},
						ExpectedUplinkChannels: []int{0, 1, 2},
						// we do not activate the CFList channels as we don't
						// now if the node knows about these frequencies
					},
					{
						Name:                   "base channels + two CFList channels are active",
						NodeChannels:           []int{0, 1, 2, 3, 4},
						ExpectedUplinkChannels: []int{0, 1, 2, 3, 4},
						// we do not activate the CFList channels as we don't
						// now if the node knows about these frequencies
					},
					{
						Name:                   "base channels + CFList are active",
						NodeChannels:           []int{0, 1, 2, 3, 4, 5, 6, 7},
						ExpectedUplinkChannels: []int{0, 1, 2, 3, 4, 5, 6, 7},
						// nothing to do, network and node are in sync
					},
					{
						Name:                   "base channels + CFList are active on node, but CFList channels are disabled on the network",
						NodeChannels:           []int{0, 1, 2, 3, 4, 5, 6, 7},
						DisabledChannels:       []int{3, 4, 5, 6, 7},
						ExpectedUplinkChannels: []int{0, 1, 2},
						ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
							{
								ChMask: lorawan.ChMask{true, true, true},
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

			Convey("Then GetUplinkChannelFrequency takes the extra channels into consideration", func() {
				tests := []int{
					868100000,
					868300000,
					868500000,
					867100000,
					867300000,
					867500000,
					867700000,
					867900000,
				}

				for expChannel, expFreq := range tests {
					var defaultChannel bool
					if expChannel < 3 {
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
							867100000,
							867300000,
							867500000,
							867700000,
							867900000,
						},
					},
				})
			})
		})
	})
}
