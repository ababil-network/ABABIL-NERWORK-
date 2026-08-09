package app

import (
	"math"
	"sync"
	"testing"
)

func TestAddBalanceOverflow(t *testing.T) {
	account := &Account{
		Balance: uint64(math.MaxUint64),
	}

	if err := AddBalance(account, 1); err == nil {
		t.Fatal("expected balance overflow error")
	}

	if account.Balance != math.MaxUint64 {
		t.Fatal("balance changed after overflow")
	}
}

func TestSubBalanceInsufficientFunds(t *testing.T) {
	account := &Account{
		Balance: 100,
	}

	if err := SubBalance(account, 101); err == nil {
		t.Fatal("expected insufficient balance error")
	}

	if account.Balance != 100 {
		t.Fatal("balance changed after failed subtraction")
	}
}

func TestReplayTryAddIsAtomic(t *testing.T) {
	replay := &ReplayManager{
		seen: make(map[string]bool),
	}

	const workers = 32

	var wg sync.WaitGroup
	successes := make(chan struct{}, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := replay.TryAdd("atomic-test-tx"); err == nil {
				successes <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(successes)

	count := 0
	for range successes {
		count++
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly one successful TryAdd, got %d",
			count,
		)
	}

	if !replay.Exists("atomic-test-tx") {
		t.Fatal("transaction was not recorded")
	}
}

func TestNonceOverflowProtection(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"

	NodeNonce.Set(address, math.MaxUint64)

	if NodeNonce.Verify(address, 0) {
		t.Fatal("nonce overflow incorrectly accepted")
	}

	if next := NodeNonce.Next(address); next != math.MaxUint64 {
		t.Fatalf(
			"expected saturated nonce %d, got %d",
			uint64(math.MaxUint64),
			next,
		)
	}
}
