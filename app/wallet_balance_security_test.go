package app

import (
	"math"
	"testing"
)

func TestCreditBalanceOverflowDoesNotWrap(t *testing.T) {
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

func TestDebitBalanceDoesNotUnderflow(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"

	original := WalletBalances
	defer func() {
		WalletBalances = original
	}()

	WalletBalances = []WalletBalance{
		{
			Address: address,
			Balance: 100,
		},
	}

	err := DebitBalance(address, 101)
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}

	if got := GetBalance(address); got != 100 {
		t.Fatalf(
			"balance changed after failed debit: got %d want %d",
			got,
			uint64(100),
		)
	}
}
