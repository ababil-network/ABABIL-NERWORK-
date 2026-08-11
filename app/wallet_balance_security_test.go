package app

import (
	"math"
	"sync"
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

func TestTransferBalanceReceiverOverflowIsAtomic(t *testing.T) {
	original := WalletBalances
	originalIndex := walletBalanceIndex

	defer func() {
		WalletBalances = original
		walletBalanceIndex = originalIndex
	}()

	from := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	to := "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: 100,
		},
		{
			Address: to,
			Balance: math.MaxUint64,
		},
	}
	walletBalanceIndex = nil

	err := TransferBalance(from, to, 50, 1)
	if err != ErrBalanceOverflow {
		t.Fatalf("expected ErrBalanceOverflow, got %v", err)
	}

	if got := GetBalance(from); got != 100 {
		t.Fatalf("sender changed after failed transfer: got %d want 100", got)
	}

	if got := GetBalance(to); got != math.MaxUint64 {
		t.Fatalf(
			"receiver changed after failed transfer: got %d want %d",
			got,
			uint64(math.MaxUint64),
		)
	}
}

func TestTransferBalanceInsufficientFundsIsAtomic(t *testing.T) {
	original := WalletBalances
	originalIndex := walletBalanceIndex

	defer func() {
		WalletBalances = original
		walletBalanceIndex = originalIndex
	}()

	from := "0xcccccccccccccccccccccccccccccccccccccc"
	to := "0xdddddddddddddddddddddddddddddddddddddd"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: 50,
		},
		{
			Address: to,
			Balance: 100,
		},
	}
	walletBalanceIndex = nil

	err := TransferBalance(from, to, 51, 51)
	if err != ErrInsufficientFunds {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}

	if got := GetBalance(from); got != 50 {
		t.Fatalf("sender changed after failed transfer: got %d want 50", got)
	}

	if got := GetBalance(to); got != 100 {
		t.Fatalf("receiver changed after failed transfer: got %d want 100", got)
	}
}

func TestTransferBalanceConcurrentIntegrity(t *testing.T) {
	original := WalletBalances
	originalIndex := walletBalanceIndex

	defer func() {
		WalletBalances = original
		walletBalanceIndex = originalIndex
	}()

	const workers = 100
	const amount = uint64(1)

	from := "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	to := "0xffffffffffffffffffffffffffffffffffffffff"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: workers,
		},
	}
	walletBalanceIndex = nil

	var wg sync.WaitGroup
	successes := make(chan struct{}, workers)
	failures := make(chan error, workers)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			if err := TransferBalance(from, to, amount, amount); err != nil {
				failures <- err
				return
			}

			successes <- struct{}{}
		}()
	}

	wg.Wait()
	close(successes)
	close(failures)

	successCount := len(successes)

	for err := range failures {
		t.Fatalf("unexpected transfer failure: %v", err)
	}

	if successCount != workers {
		t.Fatalf(
			"unexpected successful transfers: got %d want %d",
			successCount,
			workers,
		)
	}

	if got := GetBalance(from); got != 0 {
		t.Fatalf(
			"unexpected sender balance: got %d want 0",
			got,
		)
	}

	if got := GetBalance(to); got != workers {
		t.Fatalf(
			"unexpected receiver balance: got %d want %d",
			got,
			uint64(workers),
		)
	}
}
