package band

import (
	"time"

	"github.com/dsMartyn/lorawan"
)

type cn779Band struct {
	band
}

func (b *cn779Band) Name() string {
	return "CN779"
}

func (b *cn779Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     786000000,
		RX2DataRate:      0,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *cn779Band) GetDownlinkTXPower(freq uint32) int {
	return 10
}

func (b *cn779Band) GetDefaultMaxUplinkEIRP() float32 {
	return 12.15
}

func (b *cn779Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (uint32, error) {
	return 785000000, nil
}

func (b *cn779Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *cn779Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency uint32) (uint32, error) {
	return uplinkFrequency, nil
}

func (b *cn779Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newCN779Band(repeaterCompatible bool) (Band, error) {
	b := cn779Band{
		band: band{
			supportsExtraChannels: true,
			cFListMinDR:           0,
			cFListMaxDR:           5,
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
				{Frequency: 779500000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 779700000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 779900000, MinDR: 0, MaxDR: 5, enabled: true},
			},

			downlinkChannels: []Channel{
				{Frequency: 779500000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 779700000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 779900000, MinDR: 0, MaxDR: 5, enabled: true},
			},
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			LoRaWAN_1_0_0: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.0
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 250, N: 242},
					7: {M: 230, N: 222},
				},
			},
			LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 250, N: 242},
					7: {M: 230, N: 222},
				},
			},
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2A, LoRaWAN 1.0.2B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 250, N: 242},
					7: {M: 230, N: 222},
				},
			},
			LoRaWAN_1_0_3: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.3A
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 250, N: 242},
					7: {M: 230, N: 222},
				},
			},
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // RP002-1.0.0, RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
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
			LoRaWAN_1_0_0: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.0
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
			LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
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
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2A, LoRaWAN 1.0.2B
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
			LoRaWAN_1_0_3: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.3A
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
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // RP002-1.0.0, RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
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
