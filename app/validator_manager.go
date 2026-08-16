package app

import "errors"

const (
	MinimumValidatorPower = 1000
	MaximumCommission     = 20
	MaximumValidators     = 100
)

var (
	ErrValidatorRegistrationInvalid = errors.New("invalid validator registration")
)

func RegisterValidator(address string, power uint64, commission uint8) error {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	if address == "" {
		return errors.New("validator address is empty")
	}

	if !IsValidAddress(address) {
		return errors.New("invalid validator address")
	}

	if power < MinimumValidatorPower {
		return errors.New("validator power below minimum")
	}

	if commission > MaximumCommission {
		return errors.New("invalid validator commission")
	}

	if len(Validators) >= MaximumValidators {
		return errors.New("maximum validator limit reached")
	}

	for _, v := range Validators {
		if v.Address == address {
			return errors.New("validator already exists")
		}
	}

	// Validator slot is determined before mutating consensus state.
	slot := uint64(len(Validators) + 1)

	requiredCollateral, err := ValidatorDepositMicroABABILFromReferencePrice(slot)
	if err != nil {
		return err
	}

	// Secure collateral before committing validator state.
	if err := DebitBalance(address, requiredCollateral); err != nil {
		return err
	}

	if err := LockValidatorCollateral(address, slot, requiredCollateral); err != nil {
		_ = CreditBalance(address, requiredCollateral)
		return err
	}

	// The consensus mutex is already held, so use the locked helper.
	addValidatorLocked(address, power)

	Validators[len(Validators)-1].Commission = commission

	LogInfo("Validator Registered")
	LogInfo("Address : " + address)

	return nil
}

func ActiveValidatorCount() int {
	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	count := 0

	for _, v := range Validators {
		if v.Active && !v.Jailed {
			count++
		}
	}

	return count
}
