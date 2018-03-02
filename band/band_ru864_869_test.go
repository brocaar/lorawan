package band

import (
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRU864Band(t *testing.T) {
	Convey("Given the RU 864-869 band is selected", t, func() {
		band, err := GetConfig(RU_864_869, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetRX1Channel returns the uplink channel", func() {
			for i := 0; i < 3; i++ {
				rx1Chan := band.GetRX1Channel(i)
				So(rx1Chan, ShouldEqual, i)
			}
		})

		Convey("Then GetRX1Frequency returns the uplink frequency", func() {
			for _, f := range []int{868900000, 869100000} {
				freq, err := band.GetRX1Frequency(f)
				So(err, ShouldBeNil)
				So(freq, ShouldEqual, f)
			}
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
				band.AddChannel(c)
			}

			Convey("When testing GetLinkADRReqPayloadsForEnabledChannels", func() {
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
							So(band.DisableUplinkChannel(c), ShouldBeNil)
						}
						pls := band.GetLinkADRReqPayloadsForEnabledChannels(test.NodeChannels)
						So(pls, ShouldResemble, test.ExpectedLinkADRReqPayloads)

						chans, err := band.GetEnabledChannelsForLinkADRReqPayloads(test.NodeChannels, pls)
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
					channel, err := band.GetUplinkChannelNumber(expFreq)
					So(err, ShouldBeNil)
					So(channel, ShouldEqual, expChannel)
				}
			})

			Convey("Then GetCFList returns the expected CFList", func() {
				cFList := band.GetCFList()
				So(cFList, ShouldNotBeNil)
				So(*cFList, ShouldEqual, lorawan.CFList{
					864100000,
					864300000,
					864500000,
					864700000,
					864900000,
				})
			})
		})
	})
}
