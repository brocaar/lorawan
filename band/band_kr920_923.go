package band

import (
	"time"

	"github.com/dsMartyn/lorawan"
)

type kr920Band struct {
	band
}

func (b *kr920Band) Name() string {
	return "KR920"
}

func (b *kr920Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     921900000,
		RX2DataRate:      0,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *kr920Band) GetDownlinkTXPower(freq uint32) int {
	return 23
}

func (b *kr920Band) GetDefaultMaxUplinkEIRP() float32 {
	return 14
}

func (b *kr920Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (uint32, error) {
	return 923100000, nil
}

func (b *kr920Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *kr920Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency uint32) (uint32, error) {
	return uplinkFrequency, nil
}

func (b *kr920Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newKR920Band(repeaterCompatible bool) (Band, error) {
	b := kr920Band{
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
			},
			rx1DataRateTable: map[int][]int{
				0: {0, 0, 0, 0, 0, 0, 1, 2},
				1: {1, 0, 0, 0, 0, 0, 2, 3},
				2: {2, 1, 0, 0, 0, 0, 3, 4},
				3: {3, 2, 1, 0, 0, 0, 4, 5},
				4: {4, 3, 2, 1, 0, 0, 5, 5},
				5: {5, 4, 3, 2, 1, 0, 5, 7},
				6: {0, 0, 0, 0, 0, 0, 0, 0},
				7: {7, 5, 5, 4, 3, 2, 7, 7},
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
				{Frequency: 922100000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 922300000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 922500000, MinDR: 0, MaxDR: 5, enabled: true},
			},

			downlinkChannels: []Channel{
				{Frequency: 922100000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 922300000, MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: 922500000, MinDR: 0, MaxDR: 5, enabled: true},
			},
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
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
				},
			},
			LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
				},
			},
			latest: map[string]map[int]MaxPayloadSize{
				RegParamRevRP002_1_0_0: map[int]MaxPayloadSize{ // RP002-1.0.0
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
				},
				latest: map[int]MaxPayloadSize{ // RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
				},
			},
		}
	} else {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
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
				},
			},
			LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
				},
			},
			latest: map[string]map[int]MaxPayloadSize{
				RegParamRevRP002_1_0_0: map[int]MaxPayloadSize{ // RP002-1.0.0
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
				},
				latest: map[int]MaxPayloadSize{ // RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
				},
			},
		}
	}

	return &b, nil
}
