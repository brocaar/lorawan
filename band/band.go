// Package band provides band specific defaults and configuration for
// downlink communication with end-devices.
package band

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/brocaar/lorawan"
)

const latest = "latest"

// Name defines the band-name type.
type Name string

// Available protocol versions.
const (
	LoRaWAN_1_0_0 = "1.0.0"
	LoRaWAN_1_0_1 = "1.0.1"
	LoRaWAN_1_0_2 = "1.0.2"
	LoRaWAN_1_0_3 = "1.0.3"
	LoRaWAN_1_1_0 = "1.1.0"
)

// Regional parameters revisions.
const (
	RegParamRevA = "A"
	RegParamRevB = "B"
	RegParamRevC = "C"
)

// Available ISM bands (deprecated, use the common name).
const (
	AS_923     Name = "AS_923"
	AU_915_928 Name = "AU_915_928"
	CN_470_510 Name = "CN_470_510"
	CN_779_787 Name = "CN_779_787"
	EU_433     Name = "EU_433"
	EU_863_870 Name = "EU_863_870"
	IN_865_867 Name = "IN_865_867"
	KR_920_923 Name = "KR_920_923"
	US_902_928 Name = "US_902_928"
	RU_864_870 Name = "RU_864_870"
)

// Available ISM bands (by common name).
const (
	EU868 Name = "EU868"
	US915 Name = "US915"
	CN779 Name = "CN779"
	EU433 Name = "EU433"
	AU915 Name = "AU915"
	CN470 Name = "CN470"
	AS923 Name = "AS923"
	KR920 Name = "KR920"
	IN865 Name = "IN865"
	RU864 Name = "RU864"
)

// Modulation defines the modulation type.
type Modulation string

// Possible modulation types.
const (
	LoRaModulation Modulation = "LORA"
	FSKModulation  Modulation = "FSK"
)

// DataRate defines a data rate
type DataRate struct {
	uplink       bool       // data-rate can be used for uplink
	downlink     bool       // data-rate can be used for downlink
	Modulation   Modulation `json:"modulation"`
	SpreadFactor int        `json:"spreadFactor,omitempty"` // used for LoRa
	Bandwidth    int        `json:"bandwidth,omitempty"`    // in kHz, used for LoRa
	BitRate      int        `json:"bitRate,omitempty"`      // bits per second, used for FSK
}

// MaxPayloadSize defines the max payload size
type MaxPayloadSize struct {
	M int // The maximum MACPayload size length
	N int // The maximum application payload length in the absence of the optional FOpt control field
}

// Channel defines the channel structure
type Channel struct {
	Frequency int // frequency in Hz
	MinDR     int
	MaxDR     int
	enabled   bool
	custom    bool // this channel was configured by the user
}

// Defaults defines the default values defined by a band.
type Defaults struct {
	// RX2Frequency defines the fixed frequency for the RX2 receive window
	RX2Frequency int

	// RX2DataRate defines the fixed data-rate for the RX2 receive window
	RX2DataRate int

	// MaxFcntGap defines the MAC_FCNT_GAP default value.
	MaxFCntGap uint32

	// ReceiveDelay1 defines the RECEIVE_DELAY1 default value.
	ReceiveDelay1 time.Duration

	// ReceiveDelay2 defines the RECEIVE_DELAY2 default value.
	ReceiveDelay2 time.Duration

	// JoinAcceptDelay1 defines the JOIN_ACCEPT_DELAY1 default value.
	JoinAcceptDelay1 time.Duration

	// JoinAcceptDelay2 defines the JOIN_ACCEPT_DELAY2 default value.
	JoinAcceptDelay2 time.Duration
}

// Band defines the interface of a LoRaWAN band object.
type Band interface {
	// Name returns the name of the band.
	Name() string

	// GetDataRateIndex returns the index for the given data-rate parameters.
	GetDataRateIndex(uplink bool, dataRate DataRate) (int, error)

	// GetDataRate returns the data-rate for the given index.
	GetDataRate(dr int) (DataRate, error)

	// GetMaxPayloadSizeForDataRateIndex returns the max-payload size for the
	// given data-rate index, protocol version and regional-parameters revision.
	// The protocol-version and regional-parameters revision must be given
	// to make sure the maximum payload size is not exceeded when communicating
	// with a device implementing a less recent revision (which could cause
	// the device to reject the payload).
	// When the version or revision is unknown, it will return the most recent
	// implemented revision values.
	GetMaxPayloadSizeForDataRateIndex(protocolVersion, regParamRevision string, dr int) (MaxPayloadSize, error)

	// GetRX1DataRateIndex returns the RX1 data-rate given the uplink data-rate
	// and RX1 data-rate offset.
	GetRX1DataRateIndex(uplinkDR, rx1DROffset int) (int, error)

	// GetTXPowerOffset returns the TX Power offset for the given offset
	// index.
	GetTXPowerOffset(txPower int) (int, error)

	// AddChannel adds an extra (user-configured) uplink / downlink channel.
	// Note: this is not supported by every region.
	AddChannel(frequency, minDR, maxDR int) error

	// GetUplinkChannel returns the uplink channel for the given index.
	GetUplinkChannel(channel int) (Channel, error)

	// GetUplinkChannelIndex returns the uplink channel index given a frequency.
	// As it is possible that the same frequency occurs twice (eg. one time as
	// a default LoRaWAN channel and one time as a custom channel using a 250 kHz
	// data-rate), a bool must be given indicating this is a default channel.
	GetUplinkChannelIndex(frequency int, defaultChannel bool) (int, error)

	// GetDownlinkChannel returns the downlink channel for the given index.
	GetDownlinkChannel(channel int) (Channel, error)

	// DisableUplinkChannelIndex disables the given uplink channel index.
	DisableUplinkChannelIndex(channel int) error

	// EnableUplinkChannelIndex enables the given uplink channel index.
	EnableUplinkChannelIndex(channel int) error

	// GetUplinkChannelIndices returns all available uplink channel indices.
	GetUplinkChannelIndices() []int

	// GetStandardUplinkChannelIndices returns all standard available uplink
	// channel indices.
	GetStandardUplinkChannelIndices() []int

	// GetCustomUplinkChannelIndices returns all custom uplink channels.
	GetCustomUplinkChannelIndices() []int

	// GetEnabledUplinkChannelIndices returns the enabled uplink channel indices.
	GetEnabledUplinkChannelIndices() []int

	// GetDisabledUplinkChannelIndices returns the disabled uplink channel indices.
	GetDisabledUplinkChannelIndices() []int

	// GetRX1ChannelIndexForUplinkChannelIndex returns the channel to use for RX1
	// given the uplink channel index.
	GetRX1ChannelIndexForUplinkChannelIndex(uplinkChannel int) (int, error)

	// GetRX1FrequencyForUplinkFrequency returns the frequency to use for RX1
	// given the uplink frequency.
	GetRX1FrequencyForUplinkFrequency(uplinkFrequency int) (int, error)

	// GetPingSlotFrequency returns the frequency to use for the Class-B ping-slot.
	GetPingSlotFrequency(devAddr lorawan.DevAddr, beaconTime time.Duration) (int, error)

	// GetCFList returns the CFList used for OTAA activation.
	// The CFList contains the extra channels (e.g. for the EU band) or the
	// channel-mask for LoRaWAN 1.1+ devices (e.g. for the US band).
	// In case of extra channels, only the first 5 extra channels with DR 0-5
	// are returned. Other channels must be set using mac-commands. When there
	// are no extra channels, this method returns nil.
	GetCFList(protocolVersion string) *lorawan.CFList

	// GetLinkADRReqPayloadsForEnabledUplinkChannelIndices returns the LinkADRReqPayloads to
	// reconfigure the device to the current enabled channels. Note that in case of
	// activation, user-defined channels (e.g. CFList) will be ignored as it
	// is unknown if the device is aware of these extra frequencies.
	GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(deviceEnabledChannels []int) []lorawan.LinkADRReqPayload

	// GetEnabledUplinkChannelIndicesForLinkADRReqPayloads returns the enabled uplink channel
	// indices after applying the given LinkADRReqPayloads to the given enabled device
	// channels.
	GetEnabledUplinkChannelIndicesForLinkADRReqPayloads(deviceEnabledChannels []int, pls []lorawan.LinkADRReqPayload) ([]int, error)

	// GetDownlinkTXPower returns the TX power for downlink transmissions
	// using the given frequency. Depending the band, it could return different
	// values for different frequencies.
	GetDownlinkTXPower(frequency int) int

	// GetDefaultMaxUplinkEIRP returns the default uplink EIRP as defined by the
	// Regional Parameters.
	GetDefaultMaxUplinkEIRP() float32

	// GetDefaults returns the band defaults.
	GetDefaults() Defaults

	// ImplementsTXParamSetup returns if the device supports the TxParamSetup mac-command.
	ImplementsTXParamSetup(protocolVersion string) bool
}

type band struct {
	supportsExtraChannels bool
	dataRates             map[int]DataRate
	maxPayloadSizePerDR   map[string]map[string]map[int]MaxPayloadSize // LoRaWAN mac-version / Regional Parameters Revision / data-rate
	rx1DataRateTable      map[int][]int
	uplinkChannels        []Channel
	downlinkChannels      []Channel
	txPowerOffsets        []int
}

func (b *band) GetDataRateIndex(uplink bool, dataRate DataRate) (int, error) {
	for i, d := range b.dataRates {
		// some bands implement different data-rates with the same parameters
		// for uplink and downlink
		if uplink {
			if d.uplink == true && d.Modulation == dataRate.Modulation && d.Bandwidth == dataRate.Bandwidth && d.BitRate == dataRate.BitRate && d.SpreadFactor == dataRate.SpreadFactor {
				return i, nil
			}
		}
		if !uplink {
			if d.downlink == true && d.Modulation == dataRate.Modulation && d.Bandwidth == dataRate.Bandwidth && d.BitRate == dataRate.BitRate && d.SpreadFactor == dataRate.SpreadFactor {
				return i, nil
			}
		}
	}
	return 0, errors.New("lorawan/band: data-rate not found")
}

func (b *band) GetDataRate(dr int) (DataRate, error) {
	d, ok := b.dataRates[dr]
	if !ok {
		return DataRate{}, errors.New("lorawan/band: invalid data-rate")
	}

	return d, nil
}

func (b *band) GetMaxPayloadSizeForDataRateIndex(protocolVersion, regParamRevision string, dr int) (MaxPayloadSize, error) {
	regParamMap, ok := b.maxPayloadSizePerDR[protocolVersion]
	if !ok {
		regParamMap, ok = b.maxPayloadSizePerDR[latest]
		if !ok {
			return MaxPayloadSize{}, fmt.Errorf("no max payload-size for %s or latest", protocolVersion)
		}
	}

	drMap, ok := regParamMap[regParamRevision]
	if !ok {
		drMap, ok = regParamMap[latest]
		if !ok {
			return MaxPayloadSize{}, fmt.Errorf("no max-payload size for regional parameters revision %s or latest", regParamRevision)
		}
	}

	ps, ok := drMap[dr]
	if !ok {
		return MaxPayloadSize{}, errors.New("lorawan/band: invalid data-rate")
	}
	return ps, nil
}

func (b *band) GetRX1DataRateIndex(uplinkDR, rx1DROffset int) (int, error) {
	offsetSlice, ok := b.rx1DataRateTable[uplinkDR]
	if !ok {
		return 0, errors.New("lorawan/band: invalid data-rate")
	}

	if rx1DROffset > len(offsetSlice)-1 {
		return 0, errors.New("lorawan/band: invalid RX1 data-rate offset")
	}

	return offsetSlice[rx1DROffset], nil
}

func (b *band) GetTXPowerOffset(txPower int) (int, error) {
	if txPower > len(b.txPowerOffsets)-1 {
		return 0, errors.New("lorawan/band: invalid tx-power")
	}
	return b.txPowerOffsets[txPower], nil
}

func (b *band) AddChannel(frequency, minDR, maxDR int) error {
	if !b.supportsExtraChannels {
		return errors.New("lorawan/band: band does not support extra channels")
	}

	c := Channel{
		Frequency: frequency,
		MinDR:     minDR,
		MaxDR:     maxDR,
		custom:    true,
		enabled:   frequency != 0,
	}

	b.uplinkChannels = append(b.uplinkChannels, c)
	b.downlinkChannels = append(b.downlinkChannels, c)
	return nil
}

func (b *band) GetUplinkChannel(channel int) (Channel, error) {
	if channel > len(b.uplinkChannels)-1 {
		return Channel{}, errors.New("lorawan/band: invalid channel")
	}

	return b.uplinkChannels[channel], nil
}

func (b *band) GetUplinkChannelIndex(frequency int, defaultChannel bool) (int, error) {
	for i, channel := range b.uplinkChannels {
		if frequency == channel.Frequency && channel.custom != defaultChannel {
			return i, nil
		}
	}

	return 0, fmt.Errorf("lorawan/band: unknown channel for frequency: %d", frequency)
}

func (b *band) GetDownlinkChannel(channel int) (Channel, error) {
	if channel > len(b.downlinkChannels)-1 {
		return Channel{}, errors.New("lorawan/band: invalid channel")
	}
	return b.downlinkChannels[channel], nil
}

func (b *band) DisableUplinkChannelIndex(channel int) error {
	if channel > len(b.uplinkChannels)-1 {
		return errors.New("lorawan/band: channel does not exist")
	}
	b.uplinkChannels[channel].enabled = false
	return nil
}

func (b *band) EnableUplinkChannelIndex(channel int) error {
	if channel > len(b.uplinkChannels)-1 {
		return errors.New("lorawan/band: channel does not exist")
	}
	b.uplinkChannels[channel].enabled = true
	return nil
}

func (b *band) GetUplinkChannelIndices() []int {
	var out []int
	for i := range b.uplinkChannels {
		out = append(out, i)
	}
	return out
}

func (b *band) GetStandardUplinkChannelIndices() []int {
	var out []int
	for i, c := range b.uplinkChannels {
		if !c.custom {
			out = append(out, i)
		}
	}
	return out
}

func (b *band) GetCustomUplinkChannelIndices() []int {
	var out []int
	for i, c := range b.uplinkChannels {
		if c.custom {
			out = append(out, i)
		}
	}
	return out
}

func (b *band) GetEnabledUplinkChannelIndices() []int {
	var out []int
	for i, c := range b.uplinkChannels {
		if c.enabled {
			out = append(out, i)
		}
	}
	return out
}

func (b *band) GetDisabledUplinkChannelIndices() []int {
	var out []int
	for i, c := range b.uplinkChannels {
		if !c.enabled {
			out = append(out, i)
		}
	}
	return out
}

func (b *band) GetCFList(protocolVersion string) *lorawan.CFList {
	// Sending the channel-mask in the CFList is supported since LoRaWAN 1.0.3.
	// For earlier versions, only a CFList with (extra) channel-list is
	// supported.
	if !b.supportsExtraChannels && (protocolVersion == LoRaWAN_1_0_0 || protocolVersion == LoRaWAN_1_0_1 || protocolVersion == LoRaWAN_1_0_2) {
		return nil
	}

	if b.supportsExtraChannels {
		return b.getCFListChannels()
	}

	return b.getCFListChannelMask()
}

func (b *band) getCFListChannelMask() *lorawan.CFList {
	var pl lorawan.CFListChannelMaskPayload
	var chMask lorawan.ChMask

	for i, c := range b.uplinkChannels {
		if i != 0 && i%len(chMask) == 0 {
			pl.ChannelMasks = append(pl.ChannelMasks, chMask)
			chMask = lorawan.ChMask{}
		}
		chMask[i%len(chMask)] = c.enabled
	}
	pl.ChannelMasks = append(pl.ChannelMasks, chMask)

	return &lorawan.CFList{
		CFListType: lorawan.CFListChannelMask,
		Payload:    &pl,
	}
}

func (b *band) getCFListChannels() *lorawan.CFList {
	var pl lorawan.CFListChannelPayload

	var i int
	for _, c := range b.uplinkChannels {
		if c.custom && i < len(pl.Channels) && c.MinDR == 0 && c.MaxDR == 5 {
			pl.Channels[i] = uint32(c.Frequency)
			i++
		}
	}

	if pl.Channels[0] == 0 {
		return nil
	}

	return &lorawan.CFList{
		CFListType: lorawan.CFListChannel,
		Payload:    &pl,
	}
}

func (b *band) GetLinkADRReqPayloadsForEnabledUplinkChannelIndices(deviceEnabledChannels []int) []lorawan.LinkADRReqPayload {
	enabledChannels := b.GetEnabledUplinkChannelIndices()

	diff := intSliceDiff(deviceEnabledChannels, enabledChannels)
	var filteredDiff []int

	for _, c := range diff {
		if channelIsActive(deviceEnabledChannels, c) || !b.uplinkChannels[c].custom {
			filteredDiff = append(filteredDiff, c)
		}
	}

	// nothing to do
	if len(diff) == 0 || len(filteredDiff) == 0 {
		return nil
	}

	// make sure we're dealing with a sorted slice
	sort.Ints(diff)

	var payloads []lorawan.LinkADRReqPayload
	chMaskCntl := -1

	// loop over the channel blocks that contain different channels
	// note that each payload holds 16 channels and that the chMaskCntl
	// defines the block
	for _, c := range diff {
		if c/16 != chMaskCntl {
			chMaskCntl = c / 16
			pl := lorawan.LinkADRReqPayload{
				Redundancy: lorawan.Redundancy{
					ChMaskCntl: uint8(chMaskCntl),
				},
			}

			// set enabled channels in this block to active
			// note that we don't enable user defined channels (CFList) as
			// we have no knowledge if the nodes has been provisioned with
			// these frequencies
			for _, ec := range enabledChannels {
				if (!b.uplinkChannels[ec].custom || channelIsActive(deviceEnabledChannels, ec)) && ec >= chMaskCntl*16 && ec < (chMaskCntl+1)*16 {
					pl.ChMask[ec%16] = true
				}
			}

			payloads = append(payloads, pl)
		}
	}

	return payloads
}

func (b *band) GetEnabledUplinkChannelIndicesForLinkADRReqPayloads(deviceEnabledChannels []int, pls []lorawan.LinkADRReqPayload) ([]int, error) {
	chMask := make([]bool, len(b.uplinkChannels))
	for _, c := range deviceEnabledChannels {
		// make sure that we don't exceed the chMask length. in case we exceed
		// we ignore the channel as it might have been removed from the network
		if c < len(chMask) {
			chMask[c] = true
		}
	}

	for _, pl := range pls {
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

	// turn the chMask into a slice of enabled channel numbers
	var out []int
	for i, enabled := range chMask {
		if enabled {
			out = append(out, i)
		}
	}

	return out, nil
}

// y that are not in x.
func intSliceDiff(x, y []int) []int {
	var out []int

	for _, cX := range x {
		found := false
		for _, cY := range y {
			if cX == cY {
				found = true
				break
			}
		}
		if !found {
			out = append(out, cX)
		}
	}

	for _, cY := range y {
		found := false
		for _, cX := range x {
			if cY == cX {
				found = true
				break
			}
		}
		if !found {
			out = append(out, cY)
		}
	}

	return out
}

func channelIsActive(channels []int, i int) bool {
	for _, c := range channels {
		if i == c {
			return true
		}
	}
	return false
}

// GetConfig returns the band configuration for the given band.
// Please refer to the LoRaWAN specification for more details about the effect
// of the repeater and dwell time arguments.
func GetConfig(name Name, repeaterCompatible bool, dt lorawan.DwellTime) (Band, error) {
	switch name {
	case AS_923, AS923:
		return newAS923Band(repeaterCompatible, dt)
	case AU_915_928, AU915:
		return newAU915Band(repeaterCompatible, dt)
	case CN_470_510, CN470:
		return newCN470Band(repeaterCompatible)
	case CN_779_787, CN779:
		return newCN779Band(repeaterCompatible)
	case EU_433, EU433:
		return newEU433Band(repeaterCompatible)
	case EU_863_870, EU868:
		return newEU863Band(repeaterCompatible)
	case IN_865_867, IN865:
		return newIN865Band(repeaterCompatible)
	case KR_920_923, KR920:
		return newKR920Band(repeaterCompatible)
	case US_902_928, US915:
		return newUS902Band(repeaterCompatible)
	case RU_864_870, RU864:
		return newRU864Band(repeaterCompatible)
	default:
		return nil, fmt.Errorf("lorawan/band: band %s is undefined", name)
	}
}
