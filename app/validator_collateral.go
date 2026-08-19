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

// reconcileValidatorCollateralsLocked recalculates security collateral for
// every active validator according to its current dynamic slot.
//
// validatorStateMu must already be held by the caller.
//
// Permanent validator IDs are never changed by this operation.
func reconcileValidatorCollateralsLocked() error {
	type adjustment struct {
		validator string
		oldAmount uint64
		newAmount uint64
		slot      uint64
	}

	adjustments := make([]adjustment, 0, len(Validators))

	// Low-level AddValidator/legacy state may legitimately exist without
	// collateral records. In that state there is nothing to reconcile.
	//
	// Production registration always creates a collateral record before the
	// validator becomes part of the registered validator set. Therefore an
	// entirely empty collateral registry is a valid no-op, while a partially
	// populated registry remains strictly validated below.
	validatorCollateralMu.RLock()
	if len(ValidatorCollaterals) == 0 {
		validatorCollateralMu.RUnlock()
		return nil
	}

	for _, validator := range Validators {
		if !validator.Active || validator.Jailed {
			continue
		}

		required, err := ValidatorDepositMicroABABILFromReferencePrice(validator.Slot)
		if err != nil {
			validatorCollateralMu.RUnlock()
			return err
		}

		found := false

		for _, collateral := range ValidatorCollaterals {
			if collateral.Validator == validator.Address {
				adjustments = append(adjustments, adjustment{
					validator: validator.Address,
					oldAmount: collateral.Amount,
					newAmount: required,
					slot:      validator.Slot,
				})
				found = true
				break
			}
		}

		if !found {
			validatorCollateralMu.RUnlock()
			return ErrValidatorCollateralNotFound
		}
	}

	validatorCollateralMu.RUnlock()

	// Balance state must be locked for the complete validation + mutation
	// sequence so another concurrent balance operation cannot invalidate
	// the preflight checks.
	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	ensureWalletBalanceIndexLocked()

	// Preflight every balance increase and decrease before mutating anything.
	for _, a := range adjustments {
		if a.newAmount > a.oldAmount {
			additional := a.newAmount - a.oldAmount

			index, ok := walletBalanceIndex[a.validator]
			if !ok {
				return ErrWalletNotFound
			}

			if WalletBalances[index].Balance < additional {
				return ErrInsufficientFunds
			}
		}

		if a.newAmount < a.oldAmount {
			refund := a.oldAmount - a.newAmount

			index, ok := walletBalanceIndex[a.validator]
			if !ok {
				// CreditBalance would create a wallet, but validator
				// reconciliation must never silently create missing
				// validator wallet state.
				return ErrWalletNotFound
			}

			if refund > ^uint64(0)-WalletBalances[index].Balance {
				return ErrBalanceOverflow
			}
		}
	}

	// Apply all balance changes atomically under walletBalanceMu.
	for _, a := range adjustments {
		index, ok := walletBalanceIndex[a.validator]
		if !ok {
			return ErrWalletNotFound
		}

		switch {
		case a.newAmount > a.oldAmount:
			additional := a.newAmount - a.oldAmount
			WalletBalances[index].Balance -= additional

		case a.newAmount < a.oldAmount:
			refund := a.oldAmount - a.newAmount
			WalletBalances[index].Balance += refund
		}
	}

	// Update collateral records only after every balance operation has
	// successfully passed validation and been applied.
	validatorCollateralMu.Lock()
	defer validatorCollateralMu.Unlock()

	for i := range ValidatorCollaterals {
		for _, a := range adjustments {
			if ValidatorCollaterals[i].Validator != a.validator {
				continue
			}

			ValidatorCollaterals[i].Amount = a.newAmount
			ValidatorCollaterals[i].Slot = a.slot
			break
		}
	}

	return nil
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
