package band

import (
	"time"

	"github.com/brocaar/lorawan"
)

type exp24GHzBand struct {
	band
}

func (b *exp24GHzBand) Name() string {
	return "EXP24GHZ"
}

func (b *exp24GHzBand) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     2480000000,
		RX2DataRate:      3,
		MaxFCntGap:       16384,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *exp24GHzBand) GetDownlinkTXPower(freq int) int {
	return 10
}

func (b *exp24GHzBand) GetDefaultMaxUplinkEIRP() float32 {
	return 16
}

func (b *exp24GHzBand) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (int, error) {
	return 0, nil
}

func (b *exp24GHzBand) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *exp24GHzBand) GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error) {
	return uplinkFrequency, nil
}

func (b *exp24GHzBand) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newExp24GHzBand() (Band, error) {
	b := exp24GHzBand{
		band: band{
			supportsExtraChannels: true,
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 800, uplink: true, downlink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 800, uplink: true, downlink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 800, uplink: true, downlink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 800, uplink: true, downlink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 800, uplink: true, downlink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 800, uplink: true, downlink: true},
				6: {Modulation: LoRaModulation, SpreadFactor: 6, Bandwidth: 800, uplink: true, downlink: true},
				7: {Modulation: LoRaModulation, SpreadFactor: 5, Bandwidth: 800, uplink: true, downlink: true},
			},
			rx1DataRateTable: map[int][]int{
				0: {0, 0, 0, 0, 0, 0},
				1: {1, 0, 0, 0, 0, 0},
				2: {2, 1, 0, 0, 0, 0},
				3: {3, 2, 1, 0, 0, 0},
				4: {4, 3, 2, 1, 0, 0},
				5: {5, 4, 3, 2, 1, 0},
				6: {6, 5, 4, 3, 2, 1},
				7: {7, 6, 5, 4, 3, 2},
			},
			txPowerOffsets: []int{
				0,
				-2,
				-4,
				-6,
				-8,
				-10,
				-12,
				-14,
			},
			uplinkChannels: []Channel{
				{Frequency: 2474000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2476000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2478000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2480000000, MinDR: 0, MaxDR: 7, enabled: true},
			},
			downlinkChannels: []Channel{
				{Frequency: 2474000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2476000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2478000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2480000000, MinDR: 0, MaxDR: 7, enabled: true},
			},
			maxPayloadSizePerDR: map[string]map[string]map[int]MaxPayloadSize{
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{
						0: {M: 255, N: 247},
						1: {M: 255, N: 247},
						2: {M: 255, N: 247},
						3: {M: 255, N: 247},
						4: {M: 255, N: 247},
						5: {M: 255, N: 247},
						6: {M: 255, N: 247},
						7: {M: 255, N: 247},
					},
				},
			},
		},
	}

	return &b, nil
}
