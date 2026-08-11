package app

import (
	"errors"
	"math"
	"sync"
)

const (
	// Final ABABIL fee policy.
	//
	// 0-50% load:
	//     50,000 transactions per $1 equivalent.
	//
	// 50-90% load:
	//     smoothly increases from 50,000 to 20,000 TX/$1.
	//
	// 90-100% load:
	//     congestion surcharge increases the effective cost
	//     from 20,000 toward 10,000 TX/$1.

	FeeLoadStartPercent uint64 = 50
	FeeLoadMidPercent   uint64 = 90
	FeeLoadMaxPercent   uint64 = 100

	FeeTargetTXPerUSDLowLoad uint64 = 50000
	FeeTargetTXPerUSDMidLoad uint64 = 20000
	FeeTargetTXPerUSDMaxLoad uint64 = 10000
)

var (
	ErrInvalidLoad            = errors.New("invalid network load")
	ErrInvalidReferencePrice  = errors.New("invalid ABABIL reference price")
	ErrFeeCalculationOverflow = errors.New("fee calculation overflow")
)

// DynamicFeeManager contains the deterministic network-fee state.
//
// PendingTx is retained temporarily for compatibility with the existing
// mempool integration. It must not be interpreted as a gas-price tier.
type DynamicFeeManager struct {
	mu sync.RWMutex

	PendingTx uint64

	// LoadPercent is the validated network load used by the fee engine.
	LoadPercent uint64
}

var NodeDynamicFee = &DynamicFeeManager{}

// SetPending updates the current pending transaction count.
func (d *DynamicFeeManager) SetPending(count uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.PendingTx = count
}

// SetLoadPercent sets the deterministic network-load percentage.
//
// Valid range: 0-100.
func (d *DynamicFeeManager) SetLoadPercent(load uint64) error {
	if load > FeeLoadMaxPercent {
		return ErrInvalidLoad
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.LoadPercent = load

	return nil
}

// Load returns the currently configured network load.
func (d *DynamicFeeManager) Load() uint64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.LoadPercent
}

// TargetTransactionsPerUSD returns the final fee target expressed as
// transactions per one USD-equivalent.
//
// The function is deterministic and does not depend on the market price
// of ABABIL.
func (d *DynamicFeeManager) TargetTransactionsPerUSD() uint64 {
	load := d.Load()

	switch {
	case load <= FeeLoadStartPercent:
		return FeeTargetTXPerUSDLowLoad

	case load <= FeeLoadMidPercent:
		return interpolateDescending(
			FeeTargetTXPerUSDLowLoad,
			FeeTargetTXPerUSDMidLoad,
			load-FeeLoadStartPercent,
			FeeLoadMidPercent-FeeLoadStartPercent,
		)

	default:
		return interpolateDescending(
			FeeTargetTXPerUSDMidLoad,
			FeeTargetTXPerUSDMaxLoad,
			load-FeeLoadMidPercent,
			FeeLoadMaxPercent-FeeLoadMidPercent,
		)
	}
}

// CalculateUSDPerTransaction returns the USD-equivalent fee for one
// transaction according to the final load curve.
func (d *DynamicFeeManager) CalculateUSDPerTransaction() (uint64, uint64, error) {
	target := d.TargetTransactionsPerUSD()

	if target == 0 {
		return 0, 0, ErrInvalidLoad
	}

	// Fee is represented in micro-USD:
	// 1 USD = 1,000,000 micro-USD.
	const microUSDPerUSD uint64 = 1_000_000

	feeMicroUSD := microUSDPerUSD / target

	if feeMicroUSD == 0 {
		feeMicroUSD = 1
	}

	return feeMicroUSD, target, nil
}

// CalculateNativeFee converts the USD-equivalent fee into native ABABIL.
//
// referencePriceMicroUSD means:
//
//	price of 1 ABABIL expressed in micro-USD.
//
// Example:
//
//	1 ABABIL = $0.10
//	referencePriceMicroUSD = 100,000
//
// The calculation is integer-only and deterministic.
func (d *DynamicFeeManager) CalculateNativeFee(
	referencePriceMicroUSD uint64,
) (uint64, error) {
	if referencePriceMicroUSD == 0 {
		return 0, ErrInvalidReferencePrice
	}

	feeMicroUSD, _, err := d.CalculateUSDPerTransaction()
	if err != nil {
		return 0, err
	}

	// native ABABIL amount in micro-ABABIL:
	//
	// feeUSD / ABABILUSD
	//
	// expressed using micro-units:
	//
	// feeMicroUSD * 1e6 / referencePriceMicroUSD
	const scale uint64 = 1_000_000

	if feeMicroUSD > math.MaxUint64/scale {
		return 0, ErrFeeCalculationOverflow
	}

	numerator := feeMicroUSD * scale

	feeMicroABABIL := numerator / referencePriceMicroUSD

	if feeMicroABABIL == 0 {
		feeMicroABABIL = 1
	}

	return feeMicroABABIL, nil
}

func interpolateDescending(
	start uint64,
	end uint64,
	position uint64,
	rangeSize uint64,
) uint64 {
	if rangeSize == 0 {
		return end
	}

	if position >= rangeSize {
		return end
	}

	delta := start - end

	return start - (delta*position)/rangeSize
}
