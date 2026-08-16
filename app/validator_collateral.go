package app

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrValidatorCollateralExists       = errors.New("validator security collateral already exists")
	ErrValidatorCollateralNotFound     = errors.New("validator security collateral not found")
	ErrValidatorCollateralLocked       = errors.New("validator security collateral is locked")
	ErrValidatorCollateralInvalid      = errors.New("invalid validator security collateral")
	ErrValidatorCollateralOverflow     = errors.New("validator security collateral overflow")
	ErrValidatorCollateralInsufficient = errors.New("insufficient validator security collateral")
)

type ValidatorCollateral struct {
	Validator string
	Slot      uint64
	Amount    uint64 // smallest native ABABIL unit
	Locked    bool
	CreatedAt time.Time
}

var (
	validatorCollateralMu sync.RWMutex

	ValidatorCollaterals []ValidatorCollateral
)

// LockValidatorCollateral creates immutable locked security collateral.
//
// Amount is denominated in the smallest native ABABIL unit.
//
// The collateral is security-only. It is NOT:
//   - staking
//   - delegation
//   - investment
//   - interest-bearing
//   - passive-yield generating
//   - a reward source
func LockValidatorCollateral(
	validator string,
	slot uint64,
	amount uint64,
) error {
	if validator == "" || slot == 0 {
		return ErrValidatorCollateralInvalid
	}

	required, err := ValidatorDepositMicroABABILFromReferencePrice(slot)
	if err != nil {
		return err
	}

	if amount < required {
		return ErrValidatorCollateralInsufficient
	}

	validatorCollateralMu.Lock()
	defer validatorCollateralMu.Unlock()

	for _, collateral := range ValidatorCollaterals {
		if collateral.Validator == validator {
			return ErrValidatorCollateralExists
		}
	}

	ValidatorCollaterals = append(
		ValidatorCollaterals,
		ValidatorCollateral{
			Validator: validator,
			Slot:      slot,
			Amount:    amount,
			Locked:    true,
			CreatedAt: time.Now().UTC(),
		},
	)

	return nil
}

// ValidatorDepositMicroABABILFromReferencePrice obtains the validated
// current ABABIL reference price and converts the immutable USD requirement
// into the smallest native ABABIL unit.
//
// Slots 1-10 require zero collateral and therefore do not require a
// reference-price lookup.
func ValidatorDepositMicroABABILFromReferencePrice(slot uint64) (uint64, error) {
	requiredUSD, err := ValidatorDepositUSD(slot)
	if err != nil {
		return 0, err
	}

	if requiredUSD == 0 {
		return 0, nil
	}

	price, err := NodeReferencePrice.Price()
	if err != nil {
		return 0, err
	}

	return ValidatorDepositMicroABABIL(slot, price)
}

// GetValidatorCollateral returns a snapshot of validator collateral.
func GetValidatorCollateral(validator string) (ValidatorCollateral, error) {
	if validator == "" {
		return ValidatorCollateral{}, ErrValidatorCollateralInvalid
	}

	validatorCollateralMu.RLock()
	defer validatorCollateralMu.RUnlock()

	for _, collateral := range ValidatorCollaterals {
		if collateral.Validator == validator {
			return collateral, nil
		}
	}

	return ValidatorCollateral{}, ErrValidatorCollateralNotFound
}

// IsValidatorCollateralLocked verifies that collateral remains locked.
func IsValidatorCollateralLocked(validator string) bool {
	validatorCollateralMu.RLock()
	defer validatorCollateralMu.RUnlock()

	for _, collateral := range ValidatorCollaterals {
		if collateral.Validator == validator {
			return collateral.Locked
		}
	}

	return false
}

// ReleaseValidatorCollateral is intentionally restricted.
//
// Normal validator operation must never release collateral.
// A future protocol-defined exit/slashing state machine must perform
// release through an explicit consensus-authorized path.
func ReleaseValidatorCollateral(validator string) error {
	if validator == "" {
		return ErrValidatorCollateralInvalid
	}

	return ErrValidatorCollateralLocked
}
