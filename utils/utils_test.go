package utils

import (
	"fmt"
	"testing"
)

func TestRoundToThreeDecimalPlaces(t *testing.T) {
    longFloat := 69.6969696969

    want := 69.697
    got := RoundToThreeDecimalPlaces(longFloat)

    if got != want {
        fmt.Printf("want: %f, got: %f", want, got)
    }
}

func TestCalculateAverage(t *testing.T) {
    percents := []float64{0.0001, 0.0002, 0.0003, 0.0004}

    want := 0.000
    got := CalculateAverage(percents) 

    if got != want {
        fmt.Printf("want: %f, got: %f", want, got)
    }
}

func TestPerformUnitConversion(t *testing.T){
    dividend, divisor := uint64(420), uint64(69)

    want := 6.087
    got := PerformUnitConversion(dividend, divisor)

    if got != want {
        fmt.Printf("want: %f, got: %f", want, got)
    }
}
