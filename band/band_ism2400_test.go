package band

import (
	"testing"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/stretchr/testify/require"
)

func TestISM2400Band(t *testing.T) {
	assert := require.New(t)

	band, err := GetConfig(ISM2400, true, lorawan.DwellTimeNoLimit)
	assert.NoError(err)

	t.Run("GetDefaults", func(t *testing.T) {
		assert := require.New(t)
		assert.Equal(Defaults{
			RX2Frequency:     2423000000,
			RX2DataRate:      0,
			ReceiveDelay1:    time.Second,
			ReceiveDelay2:    time.Second * 2,
			JoinAcceptDelay1: time.Second * 5,
			JoinAcceptDelay2: time.Second * 6,
		}, band.GetDefaults())
	})

	t.Run("GetDownlinkTXPower", func(t *testing.T) {
		assert := require.New(t)
		assert.Equal(10, band.GetDownlinkTXPower(2423000000))
	})

	t.Run("GetPingSlotFrequency", func(t *testing.T) {
		assert := require.New(t)
		f, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
		assert.NoError(err)
		assert.EqualValues(2424000000, f)

	})

	t.Run("GetRX1ChannelIndexForUplinkChannelIndex", func(t *testing.T) {
		assert := require.New(t)
		i, err := band.GetRX1ChannelIndexForUplinkChannelIndex(3)
		assert.NoError(err)
		assert.Equal(3, i)
	})

	t.Run("GetRX1FrequencyForUplinkFrequency", func(t *testing.T) {
		assert := require.New(t)
		f, err := band.GetRX1FrequencyForUplinkFrequency(2425000000)
		assert.NoError(err)
		assert.EqualValues(2425000000, f)
	})

	t.Run("Five extra channels", func(t *testing.T) {
		chans := []uint32{
			2426000000,
			2427000000,
			2428000000,
			2429000000,
			2430000000,
		}

		for _, c := range chans {
			band.AddChannel(c, 0, 7)
		}

		t.Run("GetCustomUplinkChannelIndices", func(t *testing.T) {
			assert := require.New(t)
			assert.Equal([]int{3, 4, 5, 6, 7}, band.GetCustomUplinkChannelIndices())
		})

		t.Run("Test LinkADRReqPayload", func(t *testing.T) {
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

			for _, tst := range tests {
				t.Run(tst.Name, func(t *testing.T) {
					assert := require.New(t)
					for _, i := range tst.DisabledChannels {
						assert.NoError(band.DisableUplinkChannelIndex(i))
					}

					pls := band.GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(tst.NodeChannels)
					assert.Equal(tst.ExpectedLinkADRReqPayloads, pls)

					chans, err := band.GetEnabledUplinkChannelIndicesForLinkADRReqPayloads(tst.NodeChannels, pls)
					assert.NoError(err)
					assert.Equal(tst.ExpectedUplinkChannels, chans)
				})
			}
		})

		t.Run("GetUplinkChannelIndex", func(t *testing.T) {
			tests := []uint32{
				2403000000,
				2425000000,
				2479000000,
				2426000000,
				2427000000,
				2428000000,
				2429000000,
				2430000000,
			}

			for expChannel, expFreq := range tests {
				var defaultChannel bool
				if expChannel < 3 {
					defaultChannel = true
				}

				channel, err := band.GetUplinkChannelIndex(expFreq, defaultChannel)
				assert.NoError(err)
				assert.Equal(expChannel, channel)
			}
		})

		t.Run("GetUplinkChannelIndexForFrequencyDR", func(t *testing.T) {
			tests := []uint32{
				2403000000,
				2425000000,
				2479000000,
				2426000000,
				2427000000,
				2428000000,
				2429000000,
				2430000000,
			}

			for expChannel, freq := range tests {
				channel, err := band.GetUplinkChannelIndexForFrequencyDR(freq, 3)
				assert.NoError(err)
				assert.Equal(expChannel, channel)
			}
		})

		t.Run("GetCFList", func(t *testing.T) {
			assert := require.New(t)
			cFList := band.GetCFList(LoRaWAN_1_0_4)
			assert.NotNil(cFList)
			assert.EqualValues(&lorawan.CFList{
				CFListType: lorawan.CFListChannel,
				Payload: &lorawan.CFListChannelPayload{
					Channels: [5]uint32{
						2426000000,
						2427000000,
						2428000000,
						2429000000,
						2430000000,
					},
				},
			}, cFList)
		})
	})
}
