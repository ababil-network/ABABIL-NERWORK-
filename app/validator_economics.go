package app

import "errors"

// Validator economics are service-based.
//
// IMPORTANT:
// Validator deposits are security collateral only.
// They are NOT staking, investment, interest-bearing balances,
// passive-yield instruments, or reward-generating deposits.
//
// Validator rewards must come from eligible network service,
// such as consensus/block-processing work and protocol transaction fees.

const (
	ValidatorDepositTier1USD  uint64 = 0
	ValidatorDepositTier2USD  uint64 = 150
	ValidatorDepositTier3USD  uint64 = 300
	ValidatorDepositTier4USD  uint64 = 450
	ValidatorDepositTier5USD  uint64 = 600
	ValidatorDepositTier6USD  uint64 = 700
	ValidatorDepositTier7USD  uint64 = 800
	ValidatorDepositTier8USD  uint64 = 900
	ValidatorDepositTier9USD  uint64 = 1000
	ValidatorDepositTier10USD uint64 = 1500
)

var (
	ErrInvalidValidatorSlot         = errors.New("invalid validator slot")
	ErrValidatorDepositRequired     = errors.New("validator security deposit required")
	ErrValidatorDepositInsufficient = errors.New("insufficient validator security deposit")
	ErrValidatorDepositLocked       = errors.New("validator security deposit is locked")
)

// ValidatorDepositUSD returns the immutable security-collateral requirement
// for a validator slot.
//
// Slots 1-10 are genesis/free-entry slots.
// Slot 11 onward follows the finalized ABABIL security-collateral schedule.
func ValidatorDepositUSD(slot uint64) (uint64, error) {
	if slot == 0 {
		return 0, ErrInvalidValidatorSlot
	}

	switch {
	case slot <= 10:
		return ValidatorDepositTier1USD, nil
	case slot <= 20:
		return ValidatorDepositTier2USD, nil
	case slot <= 30:
		return ValidatorDepositTier3USD, nil
	case slot <= 40:
		return ValidatorDepositTier4USD, nil
	case slot <= 50:
		return ValidatorDepositTier5USD, nil
	case slot <= 60:
		return ValidatorDepositTier6USD, nil
	case slot <= 70:
		return ValidatorDepositTier7USD, nil
	case slot <= 80:
		return ValidatorDepositTier8USD, nil
	case slot <= 90:
		return ValidatorDepositTier9USD, nil
	default:
		return ValidatorDepositTier10USD, nil
	}
}

// ValidatorDepositMicroABABIL converts the USD-denominated security
// requirement into the smallest native ABABIL unit.
//
// microABABILPerUSD must be a validated reference-price observation.
// It represents the number of micro-ABABIL required for one USD.
//
// This function performs overflow-safe multiplication.
func ValidatorDepositMicroABABIL(slot uint64, microABABILPerUSD uint64) (uint64, error) {
	if microABABILPerUSD == 0 {
		return 0, ErrValidatorDepositRequired
	}

	usd, err := ValidatorDepositUSD(slot)
	if err != nil {
		return 0, err
	}

	if usd == 0 {
		return 0, nil
	}

	if usd > ^uint64(0)/microABABILPerUSD {
		return 0, errors.New("validator deposit conversion overflow")
	}

	return usd * microABABILPerUSD, nil
}

// ValidateValidatorDeposit verifies that the supplied collateral is enough
// for the validator's slot.
//
// The deposit remains locked as security collateral and must never be treated
// as a source of staking yield or passive reward.
func ValidateValidatorDeposit(
	slot uint64,
	depositMicroABABIL uint64,
	microABABILPerUSD uint64,
) error {
	required, err := ValidatorDepositMicroABABIL(
		slot,
		microABABILPerUSD,
	)
	if err != nil {
		return err
	}

	if depositMicroABABIL < required {
		return ErrValidatorDepositInsufficient
	}

	return nil
}
