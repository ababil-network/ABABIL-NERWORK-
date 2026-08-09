package app

import (
	"errors"
	"math"
)

const (
	DefaultGasLimit uint64 = 21000
	DefaultGasPrice uint64 = 1
)

func CalculateGasFee(limit, price uint64) (uint64, error) {
	if price == 0 {
		return 0, nil
	}

	if limit > math.MaxUint64/price {
		return 0, errors.New("gas fee overflow")
	}

	return limit * price, nil
}
