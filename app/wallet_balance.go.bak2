package app

import "errors"

type WalletBalance struct {
	Address string
	Balance uint64
}

var WalletBalances []WalletBalance

func GetBalance(address string) uint64 {

	for _, w := range WalletBalances {
		if w.Address == address {
			return w.Balance
		}
	}

	return 0
}

func CreditBalance(address string, amount uint64) {

	for i := range WalletBalances {

		if WalletBalances[i].Address == address {

			WalletBalances[i].Balance += amount

			return
		}
	}

	WalletBalances = append(WalletBalances, WalletBalance{
		Address: address,
		Balance: amount,
	})
}

func DebitBalance(address string, amount uint64) error {

	for i := range WalletBalances {

		if WalletBalances[i].Address == address {

			if WalletBalances[i].Balance < amount {
				return errors.New("insufficient balance")
			}

			WalletBalances[i].Balance -= amount

			return nil
		}
	}

	return errors.New("wallet not found")
}
