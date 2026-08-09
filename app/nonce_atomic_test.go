package app

import (
	"sync"
	"testing"
)

func TestNonceTrySetIsAtomic(t *testing.T) {
	address := "0x1111111111111111111111111111111111111111"

	original := NodeNonce
	defer func() {
		NodeNonce = original
	}()

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	const workers = 32

	var wg sync.WaitGroup
	successes := make(chan struct{}, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if NodeNonce.TrySet(address, 1) {
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
			"expected exactly one successful nonce reservation, got %d",
			count,
		)
	}

	if got := NodeNonce.Get(address); got != 1 {
		t.Fatalf(
			"unexpected final nonce: got %d want 1",
			got,
		)
	}
}
