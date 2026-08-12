package app

import (
	"sync"
	"testing"
)

func TestApplyTransactionConcurrentSameHashOnlyOneSucceeds(t *testing.T) {
	setupValidFeeEnvironment(t)
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay
	originalValidators := Validators
	originalRewardPool := CurrentRewardPool
	originalRewardHistory := RewardHistory
	originalTreasury := NetworkTreasury
	originalTreasuryHistory := TreasuryHistory
	originalWallet := NodeWallet

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
		Validators = originalValidators
		CurrentRewardPool = originalRewardPool
		RewardHistory = originalRewardHistory
		NetworkTreasury = originalTreasury
		TreasuryHistory = originalTreasuryHistory
		NodeWallet = originalWallet
	}()

	Validators = nil

	// Create a real sender wallet so the transaction has a valid
	// signature, public key, and matching sender address.
	sender, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create sender wallet: %v", err)
	}

	receiver, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create receiver wallet: %v", err)
	}

	NodeWallet = sender

	WalletBalances = []WalletBalance{
		{
			Address: sender.Address,
			Balance: 100000,
		},
	}

	NodeNonce = newTestNonceManager()

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	CurrentRewardPool = RewardPool{}
	RewardHistory = nil
	NetworkTreasury = Treasury{}
	TreasuryHistory = nil
	// Build one completely valid and signed transaction.
	tx := NewTransaction(
		sender.Address,
		receiver.Address,
		100,
	)

	if tx.Hash == "" {
		t.Fatal("transaction hash is empty")
	}

	if tx.Signature == "" {
		t.Fatal("transaction signature is empty")
	}

	if tx.PublicKey == "" {
		t.Fatal("transaction public key is empty")
	}

	if err := ValidateTransaction(tx); err != nil {
		t.Fatalf("generated transaction is invalid: %v", err)
	}

	const workers = 128

	var wg sync.WaitGroup
	results := make(chan error, workers)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			results <- ApplyTransaction(tx)
		}()
	}

	wg.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
		}
	}

	// Exactly one transaction with the same hash must succeed.
	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful transaction, got %d",
			successes,
		)
	}

	expectedFee, err := CalculateFinalNativeFee()
	if err != nil {
		t.Fatalf("failed to calculate expected transaction fee: %v", err)
	}

	expectedSenderBalance := uint64(100000) - tx.Amount - expectedFee

	if got := GetBalance(sender.Address); got != expectedSenderBalance {
		t.Fatalf(
			"unexpected sender balance: got %d want %d",
			got,
			expectedSenderBalance,
		)
	}

	if got := GetBalance(receiver.Address); got != 100 {
		t.Fatalf(
			"unexpected receiver balance: got %d want 100",
			got,
		)
	}

	if got := NodeNonce.Get(sender.Address); got != 1 {
		t.Fatalf(
			"unexpected nonce: got %d want 1",
			got,
		)
	}

	// The successful transaction must remain in replay protection.
	if !NodeReplay.Exists(tx.Hash) {
		t.Fatal("successful transaction was not retained in replay state")
	}
}
