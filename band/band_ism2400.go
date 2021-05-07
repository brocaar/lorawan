package band

import (
	"time"

	"github.com/brocaar/lorawan"
)

// see: https://lora-developers.semtech.com/library/tech-papers-and-guides/physical-layer-proposal-2.4ghz
type ism2400Band struct {
	band
}

func (b *ism2400Band) Name() string {
	return "ISM2400"
}

func (b *ism2400Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     2423000000,
		RX2DataRate:      0,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *ism2400Band) GetDownlinkTXPower(freq uint32) int {
	return 10
}

func (b *ism2400Band) GetDefaultMaxUplinkEIRP() float32 {
	return 10
}

func (b *ism2400Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (uint32, error) {
	return 2424000000, nil
}

func (b *ism2400Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *ism2400Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency uint32) (uint32, error) {
	return uplinkFrequency, nil
}

func (b *ism2400Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return true
}

func newISM2400Band(repeaterCompatible bool) (Band, error) {
	b := ism2400Band{
		band: band{
			supportsExtraChannels: true,
			cFListMinDR:           0,
			cFListMaxDR:           7,
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 812, uplink: true, downlink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 812, uplink: true, downlink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 812, uplink: true, downlink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 812, uplink: true, downlink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 812, uplink: true, downlink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 812, uplink: true, downlink: true},
				6: {Modulation: LoRaModulation, SpreadFactor: 6, Bandwidth: 812, uplink: true, downlink: true},
				7: {Modulation: LoRaModulation, SpreadFactor: 5, Bandwidth: 812, uplink: true, downlink: true},
			},
			rx1DataRateTable: map[int][]int{
				0: {0, 0, 0, 0, 0, 0},
				1: {1, 0, 0, 0, 0, 0},
				2: {2, 1, 0, 0, 0, 0},
				3: {3, 2, 1, 0, 0, 0},
				4: {4, 3, 2, 1, 0, 0},
				5: {5, 4, 2, 2, 1, 0},
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
				{Frequency: 2403000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2425000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2479000000, MinDR: 0, MaxDR: 7, enabled: true},
			},
			downlinkChannels: []Channel{
				{Frequency: 2403000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2425000000, MinDR: 0, MaxDR: 7, enabled: true},
				{Frequency: 2479000000, MinDR: 0, MaxDR: 7, enabled: true},
			},
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{
					0: {M: 59, N: 51},
					1: {M: 123, N: 115},
					2: {M: 228, N: 220},
					3: {M: 228, N: 220},
					4: {M: 228, N: 220},
					5: {M: 228, N: 220},
					6: {M: 228, N: 220},
					7: {M: 228, N: 220},
				},
			},
		}
	} else {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{
					0: {M: 59, N: 51},
					1: {M: 123, N: 115},
					2: {M: 248, N: 220},
					3: {M: 248, N: 240},
					4: {M: 248, N: 240},
					5: {M: 248, N: 240},
					6: {M: 248, N: 240},
					7: {M: 248, N: 240},
				},
			},
		}
	}

	return &b, nil
}
