package app

import (
	"fmt"
	"testing"
)

func BenchmarkSenderLockSameSender(b *testing.B) {
	address := "0x1111111111111111111111111111111111111111"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		shard := lockSender(address)
		unlockSender(shard)
	}
}

func BenchmarkSenderLockDifferentSenders(b *testing.B) {
	addresses := make([]string, 4096)

	for i := range addresses {
		addresses[i] = fmt.Sprintf("0x%040x", i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		shard := lockSender(addresses[i&4095])
		unlockSender(shard)
	}
}

func BenchmarkSenderLockParallel(b *testing.B) {
	addresses := make([]string, 4096)

	for i := range addresses {
		addresses[i] = fmt.Sprintf("0x%040x", i)
	}

	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0

		for pb.Next() {
			shard := lockSender(addresses[i&4095])
			unlockSender(shard)
			i++
		}
	})
}
