package band

import (
	"time"

	"github.com/brocaar/lorawan"
)

type in865Band struct {
	band
}

func (b *in865Band) Name() string {
	return "IN865"
}

func (b *in865Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     866550000,
		RX2DataRate:      2,
		MaxFCntGap:       16384,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *in865Band) GetDownlinkTXPower(freq int) int {
	return 27
}

func (b *in865Band) GetDefaultMaxUplinkEIRP() float32 {
	return 30
}

func (b *in865Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (int, error) {
	return 866550000, nil
}

func (b *in865Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *in865Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error) {
	return uplinkFrequency, nil
}

func (b *in865Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newIN865Band(repeaterCompatible bool) (Band, error) {
	b := in865Band{
		band: band{
			supportsExtraChannels: true,
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 125, uplink: true, downlink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 125, uplink: true, downlink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 125, uplink: true, downlink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 125, uplink: true, downlink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 125, uplink: true, downlink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 125, uplink: true, downlink: true},
				// 6
				7: {Modulation: FSKModulation, BitRate: 50000, uplink: true, downlink: true},
			},
			rx1DataRateTable: map[int][]int{
				0: {0, 0, 0, 0, 0, 0, 1, 2},
				1: {1, 0, 0, 0, 0, 0, 2, 3},
				2: {2, 1, 0, 0, 0, 0, 3, 4},
				3: {3, 2, 1, 0, 0, 0, 4, 5},
				4: {4, 3, 2, 1, 0, 0, 5, 5},
				5: {5, 4, 3, 2, 1, 0, 5, 5},
				// 6
				7: {7, 6, 5, 4, 3, 2, 7, 7},
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
				-16,
				-18,
				-20,
			},
			uplinkChannels: []Channel{
				{Frequency: 865062500, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 865402500, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 865985000, MinDR: 0, MaxDR: 5, enabled: true},
			},

			downlinkChannels: []Channel{
				{Frequency: 865062500, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 865402500, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 865985000, MinDR: 0, MaxDR: 5, enabled: true},
			},
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B, 1.1.0A, 1.1.0B
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
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B, 1.1.0A, 1.1.0B
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
