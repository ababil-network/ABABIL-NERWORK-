package app

import (
	"errors"
	"math"
)

func AddBalance(account *Account, amount uint64) error {

	if account.Balance > math.MaxUint64-amount {
		return errors.New("balance overflow")
	}

	account.Balance += amount

	return nil
}

func SubBalance(account *Account, amount uint64) error {
	if account.Balance < amount {
		return errors.New("insufficient balance")
	}

	account.Balance -= amount
	return nil
}
