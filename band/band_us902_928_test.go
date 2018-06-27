package band

import (
	"fmt"
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUS902Band(t *testing.T) {
	Convey("Given the US 902-928 band is selected", t, func() {
		band, err := GetConfig(US_902_928, true, lorawan.DwellTimeNoLimit)
		So(err, ShouldBeNil)

		Convey("Then GetDefaults returns the expected value", func() {
			So(band.GetDefaults(), ShouldResemble, Defaults{
				RX2Frequency:     923300000,
				RX2DataRate:      8,
				MaxFCntGap:       16384,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			})
		})

		Convey("Then GetDownlinkTXPower returns the expected value", func() {
			So(band.GetDownlinkTXPower(0), ShouldEqual, 20)
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
					ExpectedFrequency: 925700000,
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
				{Channel: 0, Frequency: 902300000, MinDR: 0, MaxDR: 3},
				{Channel: 63, Frequency: 914900000, MinDR: 0, MaxDR: 3},
				{Channel: 64, Frequency: 903000000, MinDR: 4, MaxDR: 4},
				{Channel: 71, Frequency: 914200000, MinDR: 4, MaxDR: 4},
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
				DataRate     int
				Channel      int
				ExpFrequency int
			}{
				{Frequency: 914900000, DataRate: 3, Channel: 63, ExpFrequency: 927500000},
				{Frequency: 903000000, DataRate: 4, Channel: 64, ExpFrequency: 923300000},
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

		Convey("When requesting all enabled channels", func() {
			active := band.GetEnabledUplinkChannelIndices()

			Convey("Then all channels are returned", func() {
				So(active, ShouldHaveLength, 72)
				for i, c := range active {
					So(i, ShouldEqual, c)
				}
			})

			Convey("Then no channels are disabled", func() {
				So(band.GetDisabledUplinkChannelIndices(), ShouldHaveLength, 0)
			})
		})

		Convey("When activating only channels 8 - 15", func() {
			for c := range band.GetEnabledUplinkChannelIndices() {
				So(band.DisableUplinkChannelIndex(c), ShouldBeNil)
			}
			for c := 8; c < 16; c++ {
				So(band.EnableUplinkChannelIndex(c), ShouldBeNil)
			}

			Convey("Then only channels 8 - 15 are active", func() {
				active := band.GetEnabledUplinkChannelIndices()
				So(active, ShouldHaveLength, 8)
				So(active, ShouldResemble, []int{8, 9, 10, 11, 12, 13, 14, 15})
			})

			Convey("Then the other channels are inactive", func() {
				inactive := band.GetDisabledUplinkChannelIndices()
				So(inactive, ShouldHaveLength, 72-8)
				var expected []int
				for i := 0; i < 8; i++ {
					expected = append(expected, i)
				}
				for i := 16; i < 72; i++ {
					expected = append(expected, i)
				}
				So(inactive, ShouldResemble, expected)
			})

			Convey("Then GetCFList for LoRaWAN 1.0.x returns nil", func() {
				So(band.GetCFList(LoRaWAN_1_0_2), ShouldBeNil)
			})

			Convey("Then GetCFList for LoRaWAN 1.1+ returns the channel-mask", func() {
				cFList := band.GetCFList(LoRaWAN_1_1_0)
				So(cFList, ShouldNotBeNil)
				So(cFList, ShouldResemble, &lorawan.CFList{
					CFListType: lorawan.CFListChannelMask,
					Payload: &lorawan.CFListChannelMaskPayload{
						ChannelMasks: []lorawan.ChMask{
							{
								false,
								false,
								false,
								false,
								false,
								false,
								false,
								false,
								true,
								true,
								true,
								true,
								true,
								true,
								true,
								true,
							},
							{}, // all false
							{}, // all false
							{}, // all false
							{}, // all false
						},
					},
				})
			})
		})

		Convey("Then GetDataRateIndex returns the expected data-rate index", func() {
			tests := []struct {
				DataRate   DataRate
				Uplink     bool
				ExpectedDR int
			}{
				{
					DataRate:   DataRate{Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 125},
					Uplink:     true,
					ExpectedDR: 0,
				},
				{
					DataRate:   DataRate{Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 500},
					Uplink:     false,
					ExpectedDR: 8,
				},
				{
					DataRate:   DataRate{Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 500},
					Uplink:     true,
					ExpectedDR: 4,
				},
				{
					DataRate:   DataRate{Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 500},
					Uplink:     false,
					ExpectedDR: 12,
				},
			}

			for _, t := range tests {
				dr, err := band.GetDataRateIndex(t.Uplink, t.DataRate)
				So(err, ShouldBeNil)
				So(dr, ShouldEqual, t.ExpectedDR)
			}
		})

		Convey("When testing GetLinkADRReqPayloadsForEnabledChannels", func() {
			var filteredChans []int
			for i := 8; i < 72; i++ {
				filteredChans = append(filteredChans, i)
			}

			tests := []struct {
				Name                       string
				NodeChannels               []int
				DisableChannels            []int
				EnableChannels             []int
				ExpectedUplinkChannels     []int
				ExpectedLinkADRReqPayloads []lorawan.LinkADRReqPayload
			}{
				{
					Name:                   "all channels active",
					NodeChannels:           band.GetUplinkChannelIndices(),
					ExpectedUplinkChannels: band.GetUplinkChannelIndices(),
				},
				{
					Name:                   "only activate channel 0 - 7",
					NodeChannels:           band.GetUplinkChannelIndices(),
					DisableChannels:        band.GetUplinkChannelIndices(),
					EnableChannels:         []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedUplinkChannels: []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
					},
				},
				{
					Name:                   "only activate channel 8 - 23",
					NodeChannels:           band.GetUplinkChannelIndices(),
					DisableChannels:        band.GetUplinkChannelIndices(),
					EnableChannels:         []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
					ExpectedUplinkChannels: []int{8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
						{
							ChMask:     lorawan.ChMask{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 1},
						},
					},
				},
				{
					Name:                   "only activate channel 64 - 71",
					NodeChannels:           band.GetUplinkChannelIndices(),
					DisableChannels:        band.GetUplinkChannelIndices(),
					EnableChannels:         []int{64, 65, 66, 67, 68, 69, 70, 71},
					ExpectedUplinkChannels: []int{64, 65, 66, 67, 68, 69, 70, 71},
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							ChMask:     lorawan.ChMask{true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 7},
						},
					},
				},
				{
					Name:                   "only disable channel 0 - 7",
					NodeChannels:           band.GetUplinkChannelIndices(),
					DisableChannels:        []int{0, 1, 2, 3, 4, 5, 6, 7},
					ExpectedUplinkChannels: filteredChans,
					ExpectedLinkADRReqPayloads: []lorawan.LinkADRReqPayload{
						{
							ChMask:     lorawan.ChMask{false, false, false, false, false, false, false, false, true, true, true, true, true, true, true, true},
							Redundancy: lorawan.Redundancy{ChMaskCntl: 0},
						},
					},
				},
			}

			for i, test := range tests {
				Convey(fmt.Sprintf("testing %s [%d]", test.Name, i), func() {
					for _, c := range test.DisableChannels {
						So(band.DisableUplinkChannelIndex(c), ShouldBeNil)
					}
					for _, c := range test.EnableChannels {
						So(band.EnableUplinkChannelIndex(c), ShouldBeNil)
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
