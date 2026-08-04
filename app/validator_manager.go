package app

import "errors"

const (
	MinimumValidatorPower = 1000
	MaximumCommission     = 20
        MaximumValidators = 100
)

func RegisterValidator(address string, power uint64, commission uint8) error {

	if address == "" {
	if !IsValidAddress(address) {
	return errors.New("invalid validator address")
}
	return errors.New("validator address is empty")
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

	AddValidator(address, power)

	Validators[len(Validators)-1].Commission = commission

	LogInfo("Validator Registered")
	LogInfo("Address : " + address)

	return nil
}
func ActiveValidatorCount() int {

	count := 0

	for _, v := range Validators {

		if v.Active && !v.Jailed {
			count++
		}
	}

	return count
}
