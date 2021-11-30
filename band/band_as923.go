package band

import (
	"fmt"
	"time"

	"github.com/dsMartyn/lorawan"
)

type as923Band struct {
	band
	nameSuffix      string
	dwellTime       lorawan.DwellTime
	frequencyOffset int
}

func (b *as923Band) Name() string {
	return "AS923" + b.nameSuffix
}

func (b *as923Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     uint32(923200000 + b.frequencyOffset),
		RX2DataRate:      2,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *as923Band) GetDownlinkTXPower(freq uint32) int {
	return 14
}

func (b *as923Band) GetDefaultMaxUplinkEIRP() float32 {
	return 16
}

func (b *as923Band) GetPingSlotFrequency(lorawan.DevAddr, time.Duration) (uint32, error) {
	return uint32(923400000 + b.frequencyOffset), nil
}

func (b *as923Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel, nil
}

func (b *as923Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency uint32) (uint32, error) {
	return uplinkFrequency, nil
}

func (b *as923Band) GetRX1DataRateIndex(uplinkDR, rx1DROffset int) (int, error) {
	if rx1DROffset < 0 || rx1DROffset > 7 {
		return 0, fmt.Errorf("lorawan/band: invalid RX1 data-rate offset: %d", rx1DROffset)
	}

	if uplinkDR < 0 || uplinkDR > 7 {
		return 0, fmt.Errorf("lorawan/band: invalid uplink data-rate: %d", uplinkDR)
	}

	minDR := 0
	if b.dwellTime == lorawan.DwellTime400ms {
		minDR = 2
	}

	effectiveRX1DROffset := []int{0, 1, 2, 3, 4, 5, -1, -2}[rx1DROffset]
	dr := uplinkDR - effectiveRX1DROffset

	if dr < minDR {
		dr = minDR
	}

	if dr > 5 {
		dr = 5
	}

	return dr, nil
}

func (b *as923Band) ImplementsTXParamSetup(protocolVersion string) bool {
	return true
}

func newAS923Band(repeaterCompatible bool, dt lorawan.DwellTime, frequencyOffset int, nameSuffix string) (Band, error) {
	b := as923Band{
		nameSuffix:      nameSuffix,
		frequencyOffset: frequencyOffset,
		dwellTime:       dt,
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
			rx1DataRateTable: map[int][]int{}, // implemented as function
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
			uplinkChannels: []Channel{
				{Frequency: uint32(923200000 + frequencyOffset), MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: uint32(923400000 + frequencyOffset), MinDR: 0, MaxDR: 5, enabled: true},
			},
			downlinkChannels: []Channel{
				{Frequency: uint32(923200000 + frequencyOffset), MinDR: 0, MaxDR: 5, enabled: true},
				{Frequency: uint32(923400000 + frequencyOffset), MinDR: 0, MaxDR: 5, enabled: true},
			},
		},
	}

	if dt == lorawan.DwellTime400ms {
		if repeaterCompatible {
			// repeater compatible + dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				LoRaWAN_1_0_3: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.3A
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // RP002-1.0.0, RP002-1.0.1, RP002-1.0.2
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 230, N: 222},
						6: {M: 230, N: 222},
						7: {M: 230, N: 222},
					},
				},
			}
		} else {
			// not repeater compatible + dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				LoRaWAN_1_0_3: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.3A
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // RP002-1.0.0, RP002-1.0.1, RP002-1.0.2
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
			}
		}
	} else {
		if repeaterCompatible {
			// repeater compatible + no dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
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
				LoRaWAN_1_0_3: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.3A
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
				LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
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
				latest: map[string]map[int]MaxPayloadSize{
					RegParamRevRP002_1_0_0: map[int]MaxPayloadSize{ // RP002-1.0.0
						0: {M: 59, N: 51},
						1: {M: 59, N: 51},
						2: {M: 59, N: 51},
						3: {M: 123, N: 115},
						4: {M: 230, N: 222},
						5: {M: 230, N: 222},
						6: {M: 230, N: 222},
						7: {M: 230, N: 222},
					},
					latest: map[int]MaxPayloadSize{ // RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
						0: {M: 59, N: 51},
						1: {M: 59, N: 51},
						2: {M: 123, N: 115},
						3: {M: 123, N: 115},
						4: {M: 230, N: 222},
						5: {M: 230, N: 222},
						6: {M: 230, N: 222},
						7: {M: 230, N: 222},
					},
				},
			}
		} else {
			// not repeater compatible + no dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_2: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B
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
				LoRaWAN_1_1_0: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0A, 1.1.0B
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
				RegParamRevRP002_1_0_0: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // RP002-1.0.0
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
					latest: map[int]MaxPayloadSize{ // RP002-1.0.1, RP002-1.0.2, RP002-1.0.3
						0: {M: 59, N: 51},
						1: {M: 59, N: 51},
						2: {M: 123, N: 115},
						3: {M: 123, N: 115},
						4: {M: 250, N: 242},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						7: {M: 250, N: 242},
					},
				},
			}
		}
	}

	return &b, nil
}
