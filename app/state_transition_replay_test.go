package app

import "testing"

func TestApplyTransactionRejectsReplay(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay
	originalWallet := NodeWallet
	originalPool := CurrentRewardPool
	originalHistory := RewardHistory

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
		NodeWallet = originalWallet
		CurrentRewardPool = originalPool
		RewardHistory = originalHistory
	}()

	wallet, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create test wallet: %v", err)
	}

	receiver, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create receiver wallet: %v", err)
	}

	NodeWallet = wallet

	WalletBalances = []WalletBalance{
		{
			Address: wallet.Address,
			Balance: 1000000,
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	CurrentRewardPool = RewardPool{}
	RewardHistory = nil

	tx := NewTransaction(
		wallet.Address,
		receiver.Address,
		100,
	)

	if tx.ID == "" || tx.Hash == "" {
		t.Fatal("failed to create transaction")
	}

	if tx.Signature == "" || tx.PublicKey == "" {
		t.Fatal("transaction was not signed")
	}

	before := GetBalance(wallet.Address)

	if err := ApplyTransaction(tx); err != nil {
		t.Fatalf("first transaction failed: %v", err)
	}

	afterFirst := GetBalance(wallet.Address)

	if afterFirst >= before {
		t.Fatal("sender balance did not decrease")
	}

	if GetBalance(receiver.Address) != tx.Amount {
		t.Fatalf(
			"receiver balance incorrect: got=%d want=%d",
			GetBalance(receiver.Address),
			tx.Amount,
		)
	}

	if !NodeReplay.Exists(tx.Hash) {
		t.Fatal("transaction was not recorded in replay protection")
	}

	// Same transaction must never be accepted again.
	if err := ApplyTransaction(tx); err == nil {
		t.Fatal("replayed transaction was accepted")
	}

	afterReplay := GetBalance(wallet.Address)

	if afterReplay != afterFirst {
		t.Fatalf(
			"sender balance changed after replay: before=%d after=%d",
			afterFirst,
			afterReplay,
		)
	}

	if GetBalance(receiver.Address) != tx.Amount {
		t.Fatalf(
			"receiver balance changed after replay: got=%d want=%d",
			GetBalance(receiver.Address),
			tx.Amount,
		)
	}

	if NodeNonce.Get(wallet.Address) != tx.Nonce {
		t.Fatalf(
			"nonce changed after replay: got=%d want=%d",
			NodeNonce.Get(wallet.Address),
			tx.Nonce,
		)
	}
}
