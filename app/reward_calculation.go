package app

import "math"

// calculatePercentage returns value * percentage / 100
// without allowing the intermediate multiplication to overflow uint64.
func calculatePercentage(value uint64, percentage uint64) uint64 {
	if percentage == 0 || value == 0 {
		return 0
	}

	if percentage == 100 {
		return value
	}

	quotient := value / 100
	remainder := value % 100

	result := quotient * percentage

	// remainder < 100 and percentage <= 100,
	// therefore remainder*percentage <= 9900.
	result += (remainder * percentage) / 100

	if result > math.MaxUint64 {
		return math.MaxUint64
	}

	return result
}
