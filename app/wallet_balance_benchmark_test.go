package app

import (
	"fmt"
	"testing"
)

func BenchmarkGetBalanceLinearScan(b *testing.B) {
	const accountCount = 10000

	addresses := make([]string, accountCount)
	balances := make([]WalletBalance, accountCount)

	for i := 0; i < accountCount; i++ {
		address := fmt.Sprintf("0x%040x", i)

		addresses[i] = address
		balances[i] = WalletBalance{
			Address: address,
			Balance: 1000000,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		address := addresses[i%accountCount]

		var balance uint64

		for _, wallet := range balances {
			if wallet.Address == address {
				balance = wallet.Balance
				break
			}
		}

		_ = balance
	}
}

func BenchmarkGetBalanceIndexed(b *testing.B) {
	const accountCount = 10000

	addresses := make([]string, accountCount)
	balances := make([]WalletBalance, accountCount)

	for i := 0; i < accountCount; i++ {
		address := fmt.Sprintf("0x%040x", i)

		addresses[i] = address
		balances[i] = WalletBalance{
			Address: address,
			Balance: 1000000,
		}
	}

	original := WalletBalances
	WalletBalances = balances

	b.Cleanup(func() {
		WalletBalances = original
	})

	// Force the production index to be rebuilt from this benchmark state.
	walletBalanceMu.Lock()
	rebuildWalletBalanceIndexLocked()
	walletBalanceMu.Unlock()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetBalance(addresses[i%accountCount])
	}
}
