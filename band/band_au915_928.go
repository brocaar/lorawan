package band

import (
	"encoding/binary"
	"sort"
	"time"

	"github.com/brocaar/lorawan"
)

type au915Band struct {
	band
	dwellTime lorawan.DwellTime
}

func (b *au915Band) Name() string {
	return "AU915"
}

func (b *au915Band) GetDefaults() Defaults {
	return Defaults{
		RX2Frequency:     923300000,
		RX2DataRate:      8,
		MaxFCntGap:       16384,
		ReceiveDelay1:    time.Second,
		ReceiveDelay2:    time.Second * 2,
		JoinAcceptDelay1: time.Second * 5,
		JoinAcceptDelay2: time.Second * 6,
	}
}

func (b *au915Band) GetDownlinkTXPower(freq int) int {
	return 27
}

func (b *au915Band) GetDefaultMaxUplinkEIRP() float32 {
	return 30
}

func (b *au915Band) GetPingSlotFrequency(devAddr lorawan.DevAddr, beaconTime time.Duration) (int, error) {
	downlinkChannel := (int(binary.BigEndian.Uint32(devAddr[:])) + int(beaconTime/(128*time.Second))) % 8
	return b.downlinkChannels[downlinkChannel].Frequency, nil
}

func (b *au915Band) GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error) {
	return uplinkChannel % 8, nil
}

func (b *au915Band) GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error) {
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

func (b *au915Band) GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(deviceEnabledChannels []int) []lorawan.LinkADRReqPayload {
	payloadsA := b.band.GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(deviceEnabledChannels)

	enabledChannels := b.GetEnabledUplinkChannelIndices()
	sort.Ints(enabledChannels)

	out := []lorawan.LinkADRReqPayload{
		{Redundancy: lorawan.Redundancy{ChMaskCntl: 7}}, // All 125 kHz OFF ChMask applies to channels 64 to 71
	}

	chMaskCntl := -1

	for _, c := range enabledChannels {
		// use the ChMask of the first LinkADRReqPayload, besides
		// turning off all 125 kHz this payload contains the ChMask
		// for the last block of channels.
		if c >= 64 {
			out[0].ChMask[c%16] = true
			continue
		}

		if c/16 != chMaskCntl {
			chMaskCntl = c / 16
			pl := lorawan.LinkADRReqPayload{
				Redundancy: lorawan.Redundancy{
					ChMaskCntl: uint8(chMaskCntl),
				},
			}

			// set the channel mask for this block
			for _, ec := range enabledChannels {
				if ec >= chMaskCntl*16 && ec < (chMaskCntl+1)*16 {
					pl.ChMask[ec%16] = true
				}
			}

			out = append(out, pl)
		}
	}

	if len(payloadsA) < len(out) {
		return payloadsA
	}
	return out
}

func (b *au915Band) GetEnabledUplinkChannelIndicesForLinkADRReqPayloads(deviceEnabledChannels []int, pls []lorawan.LinkADRReqPayload) ([]int, error) {
	chMask := make([]bool, len(b.uplinkChannels))
	for _, c := range deviceEnabledChannels {
		// make sure that we don't exceed the chMask length. in case we exceed
		// we ignore the channel as it might have been removed from the network
		if c < len(chMask) {
			chMask[c] = true
		}
	}

	for _, pl := range pls {
		if pl.Redundancy.ChMaskCntl == 6 || pl.Redundancy.ChMaskCntl == 7 {
			for i := 0; i < 64; i++ {
				if pl.Redundancy.ChMaskCntl == 6 {
					chMask[i] = true
				} else {
					chMask[i] = false
				}
			}

			for i, cm := range pl.ChMask[0:8] {
				chMask[64+i] = cm
			}
		} else {
			for i, enabled := range pl.ChMask {
				if int(pl.Redundancy.ChMaskCntl*16)+i >= len(chMask) && !enabled {
					continue
				}

				if int(pl.Redundancy.ChMaskCntl*16)+i >= len(chMask) {
					return nil, ErrChannelDoesNotExist
				}

				chMask[int(pl.Redundancy.ChMaskCntl*16)+i] = enabled
			}
		}
	}

	// turn the chMask into a slice of enabled channel numbers
	var out []int
	for i, enabled := range chMask {
		if enabled {
			out = append(out, i)
		}
	}

	return out, nil
}

func (b *au915Band) ImplementsTXParamSetup(protocolVersion string) bool {
	// In these versions it is specified that this mac-command is not implemented.
	if protocolVersion == "1.0.1" || protocolVersion == "1.0.2" {
		return false
	}

	// In later versions it is specified that this mac-command must be
	// implmented.
	return true
}

func newAU915Band(repeaterCompatible bool, dt lorawan.DwellTime) (Band, error) {
	b := au915Band{
		dwellTime: dt,
		band: band{
			dataRates: map[int]DataRate{
				0: {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 125, uplink: true},
				1: {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 125, uplink: true},
				2: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 125, uplink: true},
				3: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 125, uplink: true},
				4: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 125, uplink: true},
				5: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 125, uplink: true},
				6: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 500, uplink: true},
				// 7
				8:  {Modulation: LoRaModulation, SpreadFactor: 12, Bandwidth: 500, downlink: true},
				9:  {Modulation: LoRaModulation, SpreadFactor: 11, Bandwidth: 500, downlink: true},
				10: {Modulation: LoRaModulation, SpreadFactor: 10, Bandwidth: 500, downlink: true},
				11: {Modulation: LoRaModulation, SpreadFactor: 9, Bandwidth: 500, downlink: true},
				12: {Modulation: LoRaModulation, SpreadFactor: 8, Bandwidth: 500, downlink: true},
				13: {Modulation: LoRaModulation, SpreadFactor: 7, Bandwidth: 500, downlink: true},
			},
			rx1DataRateTable: map[int][]int{
				0: {8, 8, 8, 8, 8, 8},
				1: {9, 8, 8, 8, 8, 8},
				2: {10, 9, 8, 8, 8, 8},
				3: {11, 10, 9, 8, 8, 8},
				4: {12, 11, 10, 9, 8, 8},
				5: {13, 12, 11, 10, 9, 8},
				6: {13, 13, 12, 11, 10, 9},
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
				-16, // 8
				-18, // 9
				-20, // 10
			},
			uplinkChannels:   make([]Channel, 72),
			downlinkChannels: make([]Channel, 8),
		},
	}

	if repeaterCompatible {
		if dt == lorawan.DwellTime400ms {
			// repeater compatbile + dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						// 7
						8:  {M: 41, N: 33},
						9:  {M: 117, N: 109},
						10: {M: 230, N: 222},
						11: {M: 230, N: 222},
						12: {M: 230, N: 222},
						13: {M: 230, N: 222},
					},
				},
			}
		} else {
			// repeater compatible + no dwell time
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
						0: {M: 19, N: 11},
						1: {M: 61, N: 53},
						2: {M: 134, N: 126},
						3: {M: 250, N: 242},
						4: {M: 250, N: 242},
						// 5-7
						8:  {M: 41, N: 33},
						9:  {M: 117, N: 109},
						10: {M: 230, N: 222},
						11: {M: 230, N: 222},
						12: {M: 230, N: 222},
						13: {M: 230, N: 222},
					},
				},
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.2B, 1.1.0A, 1.1.0B
						0: {M: 59, N: 51},
						1: {M: 59, N: 51},
						2: {M: 59, N: 51},
						3: {M: 123, N: 115},
						4: {M: 230, N: 222},
						5: {M: 230, N: 222},
						6: {M: 230, N: 222},
						// 7
						8:  {M: 41, N: 33},
						9:  {M: 117, N: 109},
						10: {M: 230, N: 222},
						11: {M: 230, N: 222},
						12: {M: 230, N: 222},
						13: {M: 230, N: 222},
					},
				},
			}
		}
	} else {
		if dt == lorawan.DwellTime400ms {
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				latest: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.1.0B
						0: {M: 0, N: 0},
						1: {M: 0, N: 0},
						2: {M: 19, N: 11},
						3: {M: 61, N: 53},
						4: {M: 133, N: 125},
						5: {M: 250, N: 242},
						6: {M: 250, N: 242},
						// 7
						8:  {M: 61, N: 53},
						9:  {M: 137, N: 129},
						10: {M: 250, N: 242},
						11: {M: 250, N: 242},
						12: {M: 250, N: 242},
						13: {M: 250, N: 242},
					},
				},
			}
		} else {
			// no repeater compatibility + no dwell tiem
			b.band.maxPayloadSizePerDR = map[string]map[string]map[int]MaxPayloadSize{
				LoRaWAN_1_0_1: map[string]map[int]MaxPayloadSize{
					latest: map[int]MaxPayloadSize{ // LoRaWAN 1.0.1
						0: {M: 19, N: 11},
						1: {M: 61, N: 53},
						2: {M: 134, N: 126},
						3: {M: 250, N: 242},
						4: {M: 250, N: 242},
						// 5-7
						8:  {M: 61, N: 53},
						9:  {M: 137, N: 129},
						10: {M: 250, N: 242},
						11: {M: 250, N: 242},
						12: {M: 250, N: 242},
						13: {M: 250, N: 242},
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
						6: {M: 250, N: 242},
						// 7
						8:  {M: 61, N: 53},
						9:  {M: 137, N: 129},
						10: {M: 250, N: 242},
						11: {M: 250, N: 242},
						12: {M: 250, N: 242},
						13: {M: 250, N: 242},
					},
				},
			}
		}
	}

	// initialize uplink channel 0 - 63
	for i := 0; i < 64; i++ {
		b.uplinkChannels[i] = Channel{
			Frequency: 915200000 + (i * 200000),
			enabled:   true,
			MinDR:     0,
			MaxDR:     5,
		}
	}

	// initialize uplink channel 64 - 71
	for i := 0; i < 8; i++ {
		b.uplinkChannels[i+64] = Channel{
			Frequency: 915900000 + (i * 1600000),
			MinDR:     6,
			MaxDR:     6,
			enabled:   true,
		}
	}

	// initialize downlink channel 0 - 7
	for i := 0; i < 8; i++ {
		b.downlinkChannels[i] = Channel{
			Frequency: 923300000 + (i * 600000),
			MinDR:     8,
			MaxDR:     13,
			enabled:   true,
		}
	}

	return &b, nil
}
