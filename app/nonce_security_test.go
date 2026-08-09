package app

import (
	"sync"
	"testing"
)

func TestNonceVerifyAndSetConcurrent(t *testing.T) {
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
	accepted := make(chan struct{}, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if NodeNonce.Verify(address, 1) {
				NodeNonce.Set(address, 1)
				accepted <- struct{}{}
			}
		}()
	}

	wg.Wait()
	close(accepted)

	count := 0
	for range accepted {
		count++
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly one transaction to accept nonce 1, got %d",
			count,
		)
	}

	if NodeNonce.Get(address) != 1 {
		t.Fatalf(
			"unexpected final nonce: got %d want 1",
			NodeNonce.Get(address),
		)
	}
}
