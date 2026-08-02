package app

import "errors"

const (
	MinimumValidatorPower = 1000
	MaximumCommission     = 20
)

func RegisterValidator(address string, power uint64, commission uint8) error {

	if address == "" {
		return errors.New("validator address is empty")
	}

	if power < MinimumValidatorPower {
		return errors.New("validator power below minimum")
	}

	if commission > MaximumCommission {
		return errors.New("invalid validator commission")
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
