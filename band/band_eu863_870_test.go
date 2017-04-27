package band

import (
	"fmt"
	"testing"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEU863Band(t *testing.T) {
	Convey("Given the EU 863-870 band is selected", t, func() {
		band, err := GetConfig(EU_863_870, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetRX1Channel returns the uplink channel", func() {
			for i := 0; i < 3; i++ {
				rx1Chan := band.GetRX1Channel(i)
				So(rx1Chan, ShouldEqual, i)
			}
		})

		Convey("Then GetRX1Frequency returns the uplink frequency", func() {
			for _, f := range []int{868100000, 868200000, 868300000} {
				freq, err := band.GetRX1Frequency(f)
				So(err, ShouldBeNil)
				So(freq, ShouldEqual, f)
			}
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
				band.AddChannel(c)
			}

			Convey("When testing GetLinkADRReqPayloadsForEnabledChannels", func() {
				tests := []struct {
					Name                       string
					NodeChannels               []int
					DisabledChannels           []int
					ExpectedLinkADRReqPayloads []lorawan.LinkADRReqPayload
				}{
					{
						Name:         "no active node channels",
						NodeChannels: []int{},
						ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
							{
								ChMask: lorawan.ChMask{true, true, true},
							},
						},
						// we only activate the base channels
					},
					{
						Name:         "base channels are active",
						NodeChannels: []int{0, 1, 2},
						// we do not activate the CFList channels as we don't
						// now if the node knows about these frequencies
					},
					{
						Name:         "base channels + CFList are active",
						NodeChannels: []int{0, 1, 2, 3, 4, 5, 6, 7},
						// nothing to do, network and node are in sync
					},
					{
						Name:             "base channels + CFList are active on node, but CFList channels are disabled on the network",
						NodeChannels:     []int{0, 1, 2, 3, 4, 5, 6, 7},
						DisabledChannels: []int{3, 4, 5, 6, 7},
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
							So(band.DisableUplinkChannel(c), ShouldBeNil)
						}
						pls := band.GetLinkADRReqPayloadsForEnabledChannels(test.NodeChannels)
						So(pls, ShouldResemble, test.ExpectedLinkADRReqPayloads)
					})
				}

			})

			Convey("Then GetChannel takes the extra channels into consideration", func() {
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
					channel, err := band.GetUplinkChannelNumber(expFreq)
					So(err, ShouldBeNil)
					So(channel, ShouldEqual, expChannel)
				}
			})

			Convey("Then GetCFList returns the expected CFList", func() {
				cFList := band.GetCFList()
				So(cFList, ShouldNotBeNil)
				So(*cFList, ShouldEqual, lorawan.CFList{
					867100000,
					867300000,
					867500000,
					867700000,
					867900000,
				})
			})
		})
	})
}
