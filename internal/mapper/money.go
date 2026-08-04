package mapper

import "math"

const minorUnitsPerUnit = 100

// UnitToMinorUnits converts a decimal currency value (e.g. 19.99)
// to its minor-unit (cents) representation for storage.
func UnitToMinorUnits(value float64) int64 {
	return int64(math.Round(value * minorUnitsPerUnit))
}

// MinorUnitsToUnit converts a minor-unit (cents) value to its
// decimal currency representation for API responses.
func MinorUnitsToUnit(value int64) float64 {
	return float64(value) / minorUnitsPerUnit
}
