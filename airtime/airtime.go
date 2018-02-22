// Package airtime provides a function for calculating the time on air.
// This implements the formula as defined by:
// https://www.semtech.com/uploads/documents/LoraDesignGuide_STD.pdf.
package airtime

import (
	"errors"
	"math"
	"time"
)

// CodingRate defines the coding-rate type.
type CodingRate int

// Available coding-rates.
const (
	CodingRate45 CodingRate = 1
	CodingRate46 CodingRate = 2
	CodingRate47 CodingRate = 3
	CodingRate48 CodingRate = 4
)

// CalculateLoRaAirtime calculates the airtime for a LoRa modulated frame.
func CalculateLoRaAirtime(payloadSize, sf, bandwidth, preambleNumber int, codingRate CodingRate, headerEnabled, lowDataRateOptimization bool) (time.Duration, error) {
	symbolDuration := CalculateLoRaSymbolDuration(sf, bandwidth)
	preambleDuration := CalculateLoRaPreambleDuration(symbolDuration, preambleNumber)

	payloadSymbolNumber, err := CalculateLoRaPayloadSymbolNumber(payloadSize, sf, codingRate, headerEnabled, lowDataRateOptimization)
	if err != nil {
		return 0, err
	}

	return preambleDuration + (time.Duration(payloadSymbolNumber) * symbolDuration), nil
}

// CalculateLoRaSymbolDuration calculates the LoRa symbol duration.
func CalculateLoRaSymbolDuration(sf int, bandwidth int) time.Duration {
	return time.Duration((1 << uint(sf)) * 1000000 / bandwidth)
}

// CalculateLoRaPreambleDuration calculates the LoRa preamble duration.
func CalculateLoRaPreambleDuration(symbolDuration time.Duration, preambleNumber int) time.Duration {
	return time.Duration((100*preambleNumber)+425) * symbolDuration / 100
}

// CalculateLoRaPayloadSymbolNumber returns the number of symbols that make
// up the packet payload and header.
func CalculateLoRaPayloadSymbolNumber(payloadSize, sf int, codingRate CodingRate, headerEnabled, lowDataRateOptimization bool) (int, error) {
	var pl, spreadingFactor, h, de, cr float64

	if codingRate < 1 || codingRate > 4 {
		return 0, errors.New("codingRate must be between 1 - 4")
	}

	if lowDataRateOptimization {
		de = 1
	}

	if !headerEnabled {
		h = 1
	}

	pl = float64(payloadSize)
	spreadingFactor = float64(sf)
	cr = float64(codingRate)

	a := 8*pl - 4*spreadingFactor + 28 + 16 - 20*h
	b := 4 * (spreadingFactor - 2*de)
	c := cr + 4

	return int(8 + math.Max(math.Ceil(a/b)*c, 0)), nil
}
