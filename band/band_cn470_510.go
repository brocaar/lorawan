package band

import (
	"encoding/binary"
	"time"

	"github.com/brocaar/lorawan"
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
		MaxFCntGap:       16384,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *cn470Band) GetDownlinkTXPower(freq int) int {
	return 14
}

func (b *cn470Band) GetDefaultMaxUplinkEIRP() float32 {
	return 19.15
}

func (b *cn470Band) GetPingSlotFrequency(devAddr lorawan.DevAddr, beaconTime time.Duration) (int, error) {
	downlinkChannel := (int(binary.BigEndian.Uint32(devAddr[:])) + int(beaconTime/(128*time.Second))) % 8
	return b.downlinkChannels[downlinkChannel].Frequency, nil
}

func (b *cn470Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel % 48, nil
}

func (b *cn470Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error) {
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
				0,
				-2,
				-4,
				-6,
				-8,
				-10,
				-12,
				-14,
			},
			uplinkChannels:   make([]Channel, 96),
			downlinkChannels: make([]Channel, 48),
		},
	}

	if repeaterCompatible {
		b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1, 1.0.2B, 1.1.0A, 1.1.0B
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
			latest: map[string]map[int]MaxPayloadSize{
				latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B, 1.1.0A, 1.1.0B
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

	// initialize uplink channels
	for i := 0; i < 96; i++ {
		b.uplinkChannels[i] = Channel{
			Frequency: 470300000 + (i * 200000),
			MinDR:     0,
			MaxDR:     5,
			enabled:   true,
		}
	}

	// initialize downlink channels
	for i := 0; i < 48; i++ {
		b.downlinkChannels[i] = Channel{
			Frequency: 500300000 + (i * 200000),
			MinDR:     0,
			MaxDR:     5,
			enabled:   true,
		}
	}

	return &b, nil
}
