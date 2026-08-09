package app

import (
	"math"
	"testing"
)

func TestTransactionTotalOverflowIsRejected(t *testing.T) {
	amount := uint64(math.MaxUint64)
	fee := uint64(1)

	if amount <= math.MaxUint64-fee {
		t.Fatal("test setup did not produce uint64 overflow")
	}
}

func TestCreditBalanceOverflowIsNotSilent(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"

	original := WalletBalances
	defer func() {
		WalletBalances = original
	}()

	WalletBalances = []WalletBalance{
		{
			Address: address,
			Balance: math.MaxUint64,
		},
	}

	CreditBalance(address, 1)

	if got := GetBalance(address); got != math.MaxUint64 {
		t.Fatalf(
			"balance changed after overflow attempt: got %d want %d",
			got,
			uint64(math.MaxUint64),
		)
	}
}
