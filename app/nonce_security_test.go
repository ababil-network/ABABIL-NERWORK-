package app

import (
	"math"
	"sync"
	"testing"
)

func newTestNonceManager() *NonceManager {
	return &NonceManager{
		nonces: make(map[string]uint64),
	}
}

func TestNonceTrySetConcurrentSameNonce(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"

	original := NodeNonce
	defer func() {
		NodeNonce = original
	}()

	NodeNonce = newTestNonceManager()

	const workers = 128

	var wg sync.WaitGroup
	results := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			results <- NodeNonce.TrySet(address, 1)
		}()
	}

	wg.Wait()
	close(results)

	successes := 0

	for ok := range results {
		if ok {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful reservation, got %d",
			successes,
		)
	}

	if got := NodeNonce.Get(address); got != 1 {
		t.Fatalf(
			"unexpected final nonce: got %d want 1",
			got,
		)
	}
}

func TestNonceTrySetSequential(t *testing.T) {
	manager := newTestNonceManager()
	address := "0x2222222222222222222222222222222222222222"

	if !manager.TrySet(address, 1) {
		t.Fatal("expected nonce 1 to succeed")
	}

	if !manager.TrySet(address, 2) {
		t.Fatal("expected nonce 2 to succeed")
	}

	if !manager.TrySet(address, 3) {
		t.Fatal("expected nonce 3 to succeed")
	}

	if manager.TrySet(address, 5) {
		t.Fatal("nonce gap must be rejected")
	}

	if manager.TrySet(address, 3) {
		t.Fatal("replayed nonce must be rejected")
	}

	if got := manager.Get(address); got != 3 {
		t.Fatalf("unexpected final nonce: got %d want 3", got)
	}
}

func TestNonceRollback(t *testing.T) {
	manager := newTestNonceManager()
	address := "0x3333333333333333333333333333333333333333"

	if !manager.TrySet(address, 1) {
		t.Fatal("expected nonce 1 to succeed")
	}

	if !manager.Rollback(address, 1) {
		t.Fatal("expected rollback of nonce 1 to succeed")
	}

	if got := manager.Get(address); got != 0 {
		t.Fatalf("unexpected nonce after rollback: got %d want 0", got)
	}

	if manager.Rollback(address, 1) {
		t.Fatal("rollback of already-rolled-back nonce must be rejected")
	}

	if got := manager.Get(address); got != 0 {
		t.Fatalf("unexpected nonce after repeated rollback: got %d want 0", got)
	}
}

func TestNonceRollbackDoesNotRevertNewerNonce(t *testing.T) {
	manager := newTestNonceManager()
	address := "0x4444444444444444444444444444444444444444"

	if !manager.TrySet(address, 1) {
		t.Fatal("expected nonce 1 to succeed")
	}

	if !manager.TrySet(address, 2) {
		t.Fatal("expected nonce 2 to succeed")
	}

	// A stale rollback for nonce 1 must never revert nonce 2.
	if manager.Rollback(address, 1) {
		t.Fatal("stale rollback must be rejected")
	}

	if got := manager.Get(address); got != 2 {
		t.Fatalf(
			"newer nonce was modified by stale rollback: got %d want 2",
			got,
		)
	}
}

func TestNonceMaxUint64Protection(t *testing.T) {
	manager := newTestNonceManager()
	address := "0x5555555555555555555555555555555555555555"

	manager.Set(address, math.MaxUint64)

	if manager.TrySet(address, 0) {
		t.Fatal("nonce overflow must be rejected")
	}

	if manager.Verify(address, 0) {
		t.Fatal("nonce verification must reject overflow state")
	}

	if got := manager.Next(address); got != math.MaxUint64 {
		t.Fatalf(
			"Next must remain at MaxUint64: got %d want %d",
			got,
			uint64(math.MaxUint64),
		)
	}

	if got := manager.Get(address); got != math.MaxUint64 {
		t.Fatalf(
			"MaxUint64 state changed unexpectedly: got %d",
			got,
		)
	}
}
