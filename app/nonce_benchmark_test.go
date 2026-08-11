package app

import (
	"fmt"
	"testing"
)

func BenchmarkNonceGet(b *testing.B) {
	const accountCount = 10000

	manager := &NonceManager{
		nonces: make(map[string]uint64, accountCount),
	}

	addresses := make([]string, accountCount)

	for i := 0; i < accountCount; i++ {
		address := fmt.Sprintf("0x%040x", i)
		addresses[i] = address
		manager.nonces[address] = uint64(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = manager.Get(addresses[i%accountCount])
	}
}

func BenchmarkNonceVerify(b *testing.B) {
	const accountCount = 10000

	manager := &NonceManager{
		nonces: make(map[string]uint64, accountCount),
	}

	addresses := make([]string, accountCount)

	for i := 0; i < accountCount; i++ {
		address := fmt.Sprintf("0x%040x", i)
		addresses[i] = address
		manager.nonces[address] = uint64(i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		address := addresses[i%accountCount]
		nonce := uint64(i%accountCount) + 1

		_ = manager.Verify(address, nonce)
	}
}

func BenchmarkNonceTrySet(b *testing.B) {
	manager := &NonceManager{
		nonces: make(map[string]uint64, 1),
	}

	address := "0x0000000000000000000000000000000000000001"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.TrySet(address, uint64(i+1))
	}
}
