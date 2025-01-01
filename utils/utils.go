package utils

import (
	"fmt"
	"log"
	"strconv"
)

func RoundToThreeDecimalPlaces(value float64) float64 {
	roundedValue, err := strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	if err != nil {
		log.Printf("Error rounding value: %s", err.Error())
	}

	return roundedValue
}

func CalculateAverageUsagePercent(usagePercents []float64) float64 {
	if len(usagePercents) == 0 {
		return 0.0
	}

	var total float64
	for _, percent := range usagePercents {
		total += percent
	}

	return RoundToThreeDecimalPlaces(total / float64(len(usagePercents)))
}

func PerformUnitConversion(dividend, divisor uint64) float64 {
	result := float64(dividend) / float64(divisor)

	return RoundToThreeDecimalPlaces(result)
}
