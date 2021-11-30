package band

import (
	"encoding/binary"
	"time"

	"github.com/dsMartyn/lorawan"
)

type cn470Band struct {
	band
}

func (b *cn470Band) Name() string {
	return "CN470"
}

func (b *cn470Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     505300000,
		RX2DataRate:      0,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *cn470Band) GetDownlinkTXPower(freq uint32) int {
	return 14
}

func (b *cn470Band) GetDefaultMaxUplinkEIRP() float32 {
	return 19.15
}

func (b *cn470Band) GetPingSlotFrequency(devAddr lorawan.DevAddr, beaconTime time.Duration) (uint32, error) {
	downlinkChannel := (int(binary.BigEndian.Uint32(devAddr[:])) + int(beaconTime/(128*time.Second))) % 8
	return b.downlinkChannels[downlinkChannel].Frequency, nil
}

func (b *cn470Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel % 48, nil
}

func (b *cn470Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency uint32) (uint32, error) {
	uplinkChan, err := b.GetUplinkChannelIndex(uplinkFrequency, true)
	if err != nil {
		return 0, err
	}

	rx1Chan, err := b.GetRX1ChannelIndexForUplinkChannelIndex(uplinkChan)
	if err != nil {
		return 0, err
	}

	return b.downlinkChannels[rx1Chan].Frequency, nil
}

func (b *cn470Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return false
}

func newCN470Band(repeaterCompatible bool) (Band, error) {
	b := cn470Band{
		band: band{
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 125, uplink: true, downlink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 125, uplink: true, downlink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 125, uplink: true, downlink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 125, uplink: true, downlink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 125, uplink: true, downlink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 125, uplink: true, downlink: true},
				6: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 500, uplink: true, downlink: true},
				7: {Modulation: FSKModulation, BitRate: 50000, uplink: true, downlink: true},
			},
			rx1DataRateTable: map[int][]int{
				0: {0, 0, 0, 0, 0, 0},
				1: {1, 0, 0, 0, 0, 0},
				2: {2, 1, 0, 0, 0, 0},
				3: {3, 2, 1, 0, 0, 0},
				4: {4, 3, 2, 1, 0, 0},
				5: {5, 4, 3, 2, 1, 0},
			},
			txPowerOffsets: []int{
				0,   // 0
				-2,  // 1
				-4,  // 2
				-6,  // 3
				-8,  // 4
				-10, // 5
				-12, // 6
				-14, // 7
			},
			uplinkChannels:   make([]Channel, 96),
			downlinkChannels: make([]Channel, 48),
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
				},
			},
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2A, 1.0.2B
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
				RegParamRevRP002_1_0_1: map[int]MaxPayloadSize{
					0: {M: 0, N: 0},
					1: {M: 31, N: 23},
					2: {M: 94, N: 86},
					3: {M: 172, N: 164},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 230, N: 222},
					7: {M: 230, N: 222},
				},
				latest: map[int]MaxPayloadSize{ // RP002-1.0.2, RP002-1.0.3
					0: {M: 0, N: 0},
					1: {M: 31, N: 23},
					2: {M: 94, N: 86},
					3: {M: 192, N: 184},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
					6: {M: 230, N: 222},
					7: {M: 230, N: 222},
				},
			},
		}
	} else {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
					0: {M: 59, N: 51},
					1: {M: 59, N: 51},
					2: {M: 59, N: 51},
					3: {M: 123, N: 115},
					4: {M: 230, N: 222},
					5: {M: 230, N: 222},
				},
			},
			LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2A, 1.0.2B
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
					0: {M: 0, N: 0},
					1: {M: 31, N: 23},
					2: {M: 94, N: 86},
					3: {M: 192, N: 184},
					4: {M: 250, N: 242},
					5: {M: 250, N: 242},
					6: {M: 250, N: 242},
					7: {M: 250, N: 242},
				},
			},
		}
	}

	// initialize uplink channels
	for i := uint32(0); i < 96; i++ {
		b.uplinkChannels[i] = Channel{
			Frequency: 470300000 + (i * 200000),
			MinDR:     0,
			MaxDR:     5,
			enabled:   true,
		}
	}

	// initialize downlink channels
	for i := uint32(0); i < 48; i++ {
		b.downlinkChannels[i] = Channel{
			Frequency: 500300000 + (i * 200000),
			MinDR:     0,
			MaxDR:     5,
			enabled:   true,
		}
	}

	return &b, nil
}
