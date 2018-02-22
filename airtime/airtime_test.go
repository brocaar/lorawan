package airtime

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCalculateLoRaAirtime(t *testing.T) {
	tests := []struct {
		PayloadSize             int
		SF                      int
		Bandwidth               int
		PreambleNum             int
		CodingRate              CodingRate
		HeaderEnabled           bool
		LowDataRateOptimization bool
		ExpectedAirtime         time.Duration
	}{
		{
			PayloadSize:             13,
			SF:                      12,
			Bandwidth:               125,
			PreambleNum:             8,
			CodingRate:              CodingRate45,
			HeaderEnabled:           true,
			LowDataRateOptimization: false,
			ExpectedAirtime:         time.Duration(1155072 * 1000),
		},
	}

	Convey("Given a test-table", t, func() {
		for i, test := range tests {
			Convey(fmt.Sprintf("Test: %d", i), func() {
				d, err := CalculateLoRaAirtime(test.PayloadSize, test.SF, test.Bandwidth, test.PreambleNum, test.CodingRate, test.HeaderEnabled, test.LowDataRateOptimization)
				So(err, ShouldBeNil)
				So(d, ShouldEqual, test.ExpectedAirtime)
			})
		}
	})
}

func TestCalculateLoRaSymbolDuration(t *testing.T) {
	tests := []struct {
		SF               int
		Bandwidth        int
		ExpectedDuration time.Duration
	}{
		{
			SF:               12,
			Bandwidth:        125,
			ExpectedDuration: time.Duration(32768 * 1000),
		},
		{
			SF:               9,
			Bandwidth:        125,
			ExpectedDuration: time.Duration(4096 * 1000),
		},
		{
			SF:               9,
			Bandwidth:        500,
			ExpectedDuration: time.Duration(1024 * 1000),
		},
	}

	Convey("Given a test-table", t, func() {
		for i, test := range tests {
			Convey(fmt.Sprintf("Test: %d", i), func() {
				So(CalculateLoRaSymbolDuration(test.SF, test.Bandwidth), ShouldEqual, test.ExpectedDuration)
			})
		}
	})
}

func TestCalculateLoRaPreambleDuration(t *testing.T) {
	Convey("Given a test-table", t, func() {
		tests := []struct {
			SymbolDuration   time.Duration
			PreambleNumber   int
			ExpectedDuration time.Duration
		}{
			{
				SymbolDuration:   CalculateLoRaSymbolDuration(12, 125),
				PreambleNumber:   8,
				ExpectedDuration: time.Duration(401408 * 1000),
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Test: %d", i), func() {
				So(CalculateLoRaPreambleDuration(test.SymbolDuration, test.PreambleNumber), ShouldEqual, test.ExpectedDuration)
			})
		}
	})
}

func TestCalculateLoRaPayloadSymbolNumber(t *testing.T) {
	Convey("Given a test-table", t, func() {
		tests := []struct {
			PayloadSize             int
			SF                      int
			CodingRate              CodingRate
			HeaderEnabled           bool
			LowDataRateOptimization bool
			ExpectedNumber          int
		}{
			{
				PayloadSize:             13,
				SF:                      12,
				CodingRate:              CodingRate45,
				HeaderEnabled:           true,
				LowDataRateOptimization: false,
				ExpectedNumber:          23,
			},
			{
				PayloadSize:             13,
				SF:                      12,
				CodingRate:              CodingRate46,
				HeaderEnabled:           true,
				LowDataRateOptimization: false,
				ExpectedNumber:          26,
			},
			{
				PayloadSize:             13,
				SF:                      12,
				CodingRate:              CodingRate45,
				HeaderEnabled:           false,
				LowDataRateOptimization: false,
				ExpectedNumber:          18,
			},

			{
				PayloadSize:             50,
				SF:                      12,
				CodingRate:              CodingRate45,
				HeaderEnabled:           true,
				LowDataRateOptimization: true,
				ExpectedNumber:          58,
			},
		}

		for i, test := range tests {
			Convey(fmt.Sprintf("Test: %d", i), func() {
				num, err := CalculateLoRaPayloadSymbolNumber(test.PayloadSize, test.SF, test.CodingRate, test.HeaderEnabled, test.LowDataRateOptimization)
				So(err, ShouldBeNil)
				So(num, ShouldEqual, test.ExpectedNumber)
			})
		}
	})

}
