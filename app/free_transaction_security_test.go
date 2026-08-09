package app

import (
	"testing"
)

func TestFreeTransactionQuotaIsConsumed(t *testing.T) {
	original := NodeFreeTransaction
	defer func() {
		NodeFreeTransaction = original
	}()

	address := "0x1111111111111111111111111111111111111111"

	NodeFreeTransaction = &FreeTransactionManager{
		data: make(map[string]*FreeTransactionInfo),
	}

	if remaining := NodeFreeTransaction.Remaining(address); remaining == 0 {
		t.Fatal("expected free transaction quota")
	}

	if !NodeFreeTransaction.Use(address) {
		t.Fatal("expected first free transaction to be accepted")
	}

	remaining := NodeFreeTransaction.Remaining(address)

	if remaining == WalletFreeLimit(address) {
		t.Fatal("free transaction quota was not consumed")
	}
}

func TestApplyTransactionConsumesFreeQuota(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay
	originalFree := NodeFreeTransaction
	originalWallet := NodeWallet

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
		NodeFreeTransaction = originalFree
		NodeWallet = originalWallet
	}()

	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	receiver, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	NodeWallet = wallet

	WalletBalances = []WalletBalance{
		{
			Address: wallet.Address,
			Balance: 1_000_000,
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	NodeFreeTransaction = &FreeTransactionManager{
		data: make(map[string]*FreeTransactionInfo),
	}

	RegisterWallet(wallet.Address)

	tx := NewTransaction(wallet.Address, receiver.Address, 100)

	if tx.Fee != 0 {
		t.Fatal("expected a free transaction")
	}

	before := NodeFreeTransaction.Remaining(wallet.Address)

	if err := ApplyTransaction(tx); err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	after := NodeFreeTransaction.Remaining(wallet.Address)

	if after != before-1 {
		t.Fatalf(
			"free quota was not consumed: before=%d after=%d",
			before,
			after,
		)
	}
}
