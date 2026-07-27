package app

import "errors"

func AddBalance(account *Account, amount uint64) {
	account.Balance += amount
}

func SubBalance(account *Account, amount uint64) error {
	if account.Balance < amount {
		return errors.New("insufficient balance")
	}

	account.Balance -= amount
	return nil
}
