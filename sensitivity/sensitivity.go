// Package sensitivity provides functions for calculating the LoRa sensitivity.
package sensitivity

import (
	"math"
)

// CalculateSensitivity calculates the LoRa sensitivity.
// The bandwidth must be given in Hz!
func CalculateSensitivity(bandwidth int, noiseFigure, snr float32) float32 {
	// see also: http://www.techplayon.com/lora-link-budget-sensitivity-calculations-example-explained/
	logBW := 10 * math.Log10(float64(bandwidth))
	return float32(-174 + logBW + float64(noiseFigure+snr))
}

// CalculateLinkBudget calculates the link budget.
// The bandwidth must be given in Hz!
func CalculateLinkBudget(bandwidth int, noiseFigure, snr, txPower float32) float32 {
	// see also: http://www.techplayon.com/lora-link-budget-sensitivity-calculations-example-explained/
	return txPower - CalculateSensitivity(bandwidth, noiseFigure, snr)
}
