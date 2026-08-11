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
	ErrDuplicateWallet   = errors.New("duplicate wallet")
	ErrInvalidAddress    = errors.New("invalid wallet address")
)

type WalletBalance struct {
	Address string
	Balance uint64
}

var (
	walletBalanceMu sync.RWMutex

	// WalletBalances remains the canonical in-memory state representation
	// for the current application layer.
	//
	// The address index below is derived from this state and is used only
	// to accelerate lookups.
	WalletBalances []WalletBalance

	walletBalanceIndex map[string]int
)

func ensureWalletBalanceIndexLocked() {
	if walletBalanceIndex == nil {
		walletBalanceIndex = make(map[string]int, len(WalletBalances))

		for i := range WalletBalances {
			address := WalletBalances[i].Address

			// Keep the first occurrence as canonical if legacy/test state
			// contains duplicates. Production state must never contain them.
			if _, exists := walletBalanceIndex[address]; !exists {
				walletBalanceIndex[address] = i
			}
		}

		return
	}

	// Tests and legacy code may replace WalletBalances directly.
	// Detect that situation and rebuild the derived index.
	if len(walletBalanceIndex) != len(WalletBalances) {
		rebuildWalletBalanceIndexLocked()
		return
	}

	for address, index := range walletBalanceIndex {
		if index < 0 ||
			index >= len(WalletBalances) ||
			WalletBalances[index].Address != address {
			rebuildWalletBalanceIndexLocked()
			return
		}
	}
}

func rebuildWalletBalanceIndexLocked() {
	walletBalanceIndex = make(map[string]int, len(WalletBalances))

	for i := range WalletBalances {
		address := WalletBalances[i].Address

		if _, exists := walletBalanceIndex[address]; exists {
			continue
		}

		walletBalanceIndex[address] = i
	}
}

func GetBalance(address string) uint64 {
	walletBalanceMu.RLock()

	if walletBalanceIndex != nil {
		if index, ok := walletBalanceIndex[address]; ok {
			if index >= 0 &&
				index < len(WalletBalances) &&
				WalletBalances[index].Address == address {
				balance := WalletBalances[index].Balance
				walletBalanceMu.RUnlock()
				return balance
			}
		}
	}

	walletBalanceMu.RUnlock()

	// The index may have become stale because current tests/legacy code
	// directly replaced WalletBalances. Rebuild it under the write lock.
	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	ensureWalletBalanceIndexLocked()

	if index, ok := walletBalanceIndex[address]; ok {
		return WalletBalances[index].Balance
	}

	return 0
}

// CreditBalance atomically credits an individual wallet.
func CreditBalance(address string, amount uint64) error {
	if address == "" {
		return ErrInvalidAddress
	}

	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	ensureWalletBalanceIndexLocked()

	if index, ok := walletBalanceIndex[address]; ok {
		if amount > math.MaxUint64-WalletBalances[index].Balance {
			return ErrBalanceOverflow
		}

		WalletBalances[index].Balance += amount
		return nil
	}

	WalletBalances = append(WalletBalances, WalletBalance{
		Address: address,
		Balance: amount,
	})

	walletBalanceIndex[address] = len(WalletBalances) - 1

	return nil
}

// DebitBalance atomically debits an individual wallet.
func DebitBalance(address string, amount uint64) error {
	if address == "" {
		return ErrInvalidAddress
	}

	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	ensureWalletBalanceIndexLocked()

	index, ok := walletBalanceIndex[address]
	if !ok {
		return ErrWalletNotFound
	}

	if WalletBalances[index].Balance < amount {
		return ErrInsufficientFunds
	}

	WalletBalances[index].Balance -= amount

	return nil
}

// TransferBalance atomically debits the sender and credits the receiver.
// No partial balance mutation is allowed.
func TransferBalance(from, to string, debitAmount, creditAmount uint64) error {
	if from == "" || to == "" {
		return ErrInvalidAddress
	}

	if from == to {
		return errors.New("sender and receiver cannot be the same")
	}

	walletBalanceMu.Lock()
	defer walletBalanceMu.Unlock()

	ensureWalletBalanceIndexLocked()

	fromIndex, fromExists := walletBalanceIndex[from]
	if !fromExists {
		return ErrWalletNotFound
	}

	if WalletBalances[fromIndex].Balance < debitAmount {
		return ErrInsufficientFunds
	}

	toIndex, toExists := walletBalanceIndex[to]

	// Validate receiver overflow before changing either balance.
	if toExists &&
		creditAmount > math.MaxUint64-WalletBalances[toIndex].Balance {
		return ErrBalanceOverflow
	}

	// All validation has passed. Commit both mutations.
	WalletBalances[fromIndex].Balance -= debitAmount

	if toExists {
		WalletBalances[toIndex].Balance += creditAmount
		return nil
	}

	WalletBalances = append(WalletBalances, WalletBalance{
		Address: to,
		Balance: creditAmount,
	})

	walletBalanceIndex[to] = len(WalletBalances) - 1

	return nil
}
