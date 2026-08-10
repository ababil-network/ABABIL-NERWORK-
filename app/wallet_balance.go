package app

import (
	"errors"
	"math"
	"sync"
)

var (
	ErrBalanceOverflow   = errors.New("balance overflow")
	ErrInsufficientFunds = errors.New("insufficient balance")
	ErrWalletNotFound    = errors.New("wallet not found")
)

type WalletBalance struct {
	Address string
	Balance uint64
}

var (
	walletBalanceMu sync.RWMutex
	WalletBalances  []WalletBalance
)

func GetBalance(address string) uint64 {
	walletBalanceMu.RLock()
	defer walletBalanceMu.RUnlock()

	for _, w := range WalletBalances {
		if w.Address == address {
			return w.Balance
		}
	}

	return 0
}

// CreditBalance adds amount to an existing wallet balance.
//
// IMPORTANT:
// A failed credit must never silently look like a successful state
// transition. The function therefore returns an error on overflow.
func CreditBalance(address string, amount uint64) error {
	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	for i := range WalletBalances {
		if WalletBalances[i].Address != address {
			continue
		}

		if amount > math.MaxUint64-WalletBalances[i].Balance {
			return ErrBalanceOverflow
		}

		WalletBalances[i].Balance += amount
		return nil
	}

	// amount itself is uint64, therefore a new balance cannot overflow.
	WalletBalances = append(WalletBalances, WalletBalance{
		Address: address,
		Balance: amount,
	})

	return nil
}

func DebitBalance(address string, amount uint64) error {
	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	for i := range WalletBalances {
		if WalletBalances[i].Address != address {
			continue
		}

		if WalletBalances[i].Balance < amount {
			return ErrInsufficientFunds
		}

		WalletBalances[i].Balance -= amount
		return nil
	}

	return ErrWalletNotFound
}
