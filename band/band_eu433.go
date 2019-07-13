package band

import (
	"time"

	"github.com/brocaar/lorawan"
)

type eu443Band struct {
	band
}

func (b *eu443Band) Name() string {
	return "EU433"
}

func (b *eu443Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     434665000,
		RX2DataRate:      0,
		MaxFCntGap:       16384,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *eu443Band) GetDownlinkTXPower(freq int) int {
	return 10
}

func (b *eu443Band) GetDefaultMaxUplinkEIRP() float32 {
	return 12.15
}

func (b *eu443Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (int, error) {
	return 434665000, nil
}

func (b *eu443Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *eu443Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error) {
	return uplinkFrequency, nil
}

func (b *eu443Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newEU433Band(repeaterCompatible bool) (Band, error) {
	b := eu443Band{
		band: band{
			supportsExtraChannels: true,
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 125, uplink: true, downlink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 125, uplink: true, downlink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 125, uplink: true, downlink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 125, uplink: true, downlink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 125, uplink: true, downlink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 125, uplink: true, downlink: true},
				6: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 250, uplink: true, downlink: true},
				7: {Modulation: FSKModulation, BitRate: 50000, uplink: true, downlink: true},
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
			},
			uplinkChannels: []Channel{
				{Frequency: 433175000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 433375000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 433575000, MinDR: 0, MaxDR: 5, enabled: true},
			},

			downlinkChannels: []Channel{
				{Frequency: 433175000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 433375000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 433575000, MinDR: 0, MaxDR: 5, enabled: true},
			},
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.0, 1.0.1, 1.0.2B, 1.1.0A, 1.1.0B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 230, N: 222},
					7: {M: 230, N: 222},
				},
			},
		}
	} else {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.0, 1.0.1, 1.0.2B, 1.1.0A, 1.1.0B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
					6: {M: 250, N: 242},
					7: {M: 250, N: 242},
				},
			},
		}
	}

	return &b, nil
}
