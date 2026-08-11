package app

import (
	"errors"
	"math"
)

const (
	// Legacy compatibility value for existing test fixtures.
	// It is NOT used by the final ABABIL transaction-fee model.
	DefaultGasPrice uint64 = 1

	DefaultGasLimit uint64 = 21000

	// Reference-price unit:
	// 1 ABABIL = referencePriceMicroUSD / 1,000,000 USD.
	MicroUSDPerUnit uint64 = 1_000_000
)

var (
	ErrGasFeeOverflow = errors.New("transaction fee calculation overflow")
)

// CalculateNativeFeeFromReferencePrice converts a USD-equivalent
// transaction fee into the smallest native ABABIL unit.
//
// feeMicroUSD:
//
//	transaction fee expressed in micro-USD.
//
// referencePriceMicroUSD:
//
//	price of 1 ABABIL expressed in micro-USD.
func CalculateNativeFeeFromReferencePrice(
	feeMicroUSD uint64,
	referencePriceMicroUSD uint64,
) (uint64, error) {
	if referencePriceMicroUSD == 0 {
		return 0, ErrInvalidReferencePrice
	}

	if feeMicroUSD == 0 {
		return 0, errors.New("invalid zero USD transaction fee")
	}

	if feeMicroUSD > math.MaxUint64/MicroUSDPerUnit {
		return 0, ErrGasFeeOverflow
	}

	numerator := feeMicroUSD * MicroUSDPerUnit

	fee := numerator / referencePriceMicroUSD

	if fee == 0 {
		fee = 1
	}

	return fee, nil
}
