package band

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/brocaar/lorawan"
)

func TestAS923_1_Band(t *testing.T) {
	t.Run("400ms dwell-time", func(t *testing.T) {
		assert := require.New(t)

		band, err := GetConfig(AS923_1, true, lorawan.DwellTime400ms)
		assert.NoError(err)

		t.Run("GetDefaults", func(t *testing.T) {
			assert := require.New(t)
			assert.Equal(Defaults{
				RX2Frequency:     923200000,
				RX2DataRate:      2,
				ReceiveDelay1:    time.Second,
				ReceiveDelay2:    time.Second * 2,
				JoinAcceptDelay1: time.Second * 5,
				JoinAcceptDelay2: time.Second * 6,
			}, band.GetDefaults())
		})

		t.Run("GetDownlinkTXPower", func(t *testing.T) {
			assert := require.New(t)
			assert.Equal(14, band.GetDownlinkTXPower(0))
		})

		t.Run("GetPingSlotFrequency", func(t *testing.T) {
			assert := require.New(t)
			freq, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
			assert.NoError(err)
			assert.Equal(923400000, freq)
		})

		t.Run("GetRX1ChannelIndexForUplinkChannelIndex", func(t *testing.T) {
			assert := require.New(t)
			c, err := band.GetRX1ChannelIndexForUplinkChannelIndex(2)
			assert.NoError(err)
			assert.Equal(2, c)
		})

		t.Run("RX1FrequencyForUplinkFrequency", func(t *testing.T) {
			assert := require.New(t)
			f, err := band.GetRX1FrequencyForUplinkFrequency(923200000)
			assert.NoError(err)
			assert.Equal(923200000, f)
		})

		t.Run("GetRX1DataRateIndex", func(t *testing.T) {
			assert := require.New(t)
			tests := []struct {
				UplinkDR    int
				RX1DROffset int
				ExpectedDR  int
			}{
				{5, 0, 5},
				{5, 1, 4},
				{5, 2, 3},
				{5, 3, 2},
				{5, 4, 2},
				{5, 5, 2},
				{5, 6, 5},
				{5, 7, 5},
				{2, 6, 3},
				{2, 7, 4},
			}

			for _, tst := range tests {
				dr, err := band.GetRX1DataRateIndex(tst.UplinkDR, tst.RX1DROffset)
				assert.NoError(err)
				assert.Equal(tst.ExpectedDR, dr)
			}
		})
	})

	t.Run("No dwell-time", func(t *testing.T) {
		assert := require.New(t)

		band, err := GetConfig(AS_923, true, lorawan.DwellTimeNoLimit)
		assert.NoError(err)

		t.Run("GetRX1DataRateIndex", func(t *testing.T) {
			assert := require.New(t)

			tests := []struct {
				UplinkDR    int
				RX1DROffset int
				ExpectedDR  int
			}{
				{5, 0, 5},
				{5, 1, 4},
				{5, 2, 3},
				{5, 3, 2},
				{5, 4, 1},
				{5, 5, 0},
				{5, 6, 5},
				{5, 7, 5},
				{2, 6, 3},
				{2, 7, 4},
			}

			for _, tst := range tests {
				dr, err := band.GetRX1DataRateIndex(tst.UplinkDR, tst.RX1DROffset)
				assert.NoError(err)
				assert.Equal(tst.ExpectedDR, dr)
			}
		})
	})
}

func TestAS923_2_Band(t *testing.T) {
	assert := require.New(t)
	band, err := GetConfig(AS923_2, true, lorawan.DwellTimeNoLimit)
	assert.NoError(err)

	assert.Equal(923200000-1800000, band.GetDefaults().RX2Frequency)
	freq, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
	assert.NoError(err)
	assert.Equal(923400000-1800000, freq)

	bandd := band.(*as923Band)

	assert.Equal(923200000-1800000, bandd.uplinkChannels[0].Frequency)
	assert.Equal(923200000-1800000, bandd.downlinkChannels[0].Frequency)
	assert.Equal(923400000-1800000, bandd.uplinkChannels[1].Frequency)
	assert.Equal(923400000-1800000, bandd.downlinkChannels[1].Frequency)
}

func TestAS923_3_Band(t *testing.T) {
	assert := require.New(t)
	band, err := GetConfig(AS923_3, true, lorawan.DwellTimeNoLimit)
	assert.NoError(err)

	assert.Equal(923200000-6600000, band.GetDefaults().RX2Frequency)
	freq, err := band.GetPingSlotFrequency(lorawan.DevAddr{}, 0)
	assert.NoError(err)
	assert.Equal(923400000-6600000, freq)

	bandd := band.(*as923Band)

	assert.Equal(923200000-6600000, bandd.uplinkChannels[0].Frequency)
	assert.Equal(923200000-6600000, bandd.downlinkChannels[0].Frequency)
	assert.Equal(923400000-6600000, bandd.uplinkChannels[1].Frequency)
	assert.Equal(923400000-6600000, bandd.downlinkChannels[1].Frequency)
}
