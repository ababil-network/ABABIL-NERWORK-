package app

import "errors"

func ValidateTransaction(tx Transaction) error {
	if tx.From == "" {
		return errors.New("sender address is empty")
	}

	if tx.To == "" {
		return errors.New("receiver address is empty")
	}

	if tx.Amount == 0 {
		return errors.New("amount must be greater than zero")
	}

	return nil
}
