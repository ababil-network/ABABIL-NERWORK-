package app

import "errors"

func RegisterValidator(address string, power uint64) error {

	if address == "" {
		return errors.New("validator address is empty")
	}

	if power == 0 {
		return errors.New("validator power must be greater than zero")
	}

	AddValidator(address, power)

	LogInfo("Validator Registered")
	LogInfo("Address : " + address)

	return nil
}
