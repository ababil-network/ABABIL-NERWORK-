package app

import (
	"sync"
	"testing"
	"time"
)

func TestApplyTransactionDoesNotConsumeNonceOnValidationFailure(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
	}()

	from := "0x1111111111111111111111111111111111111111"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: 1000,
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	// Sender and receiver are intentionally identical.
	// ValidateTransaction must reject this before state mutation.
	tx := Transaction{
		ID:        "atomic-validation-failure",
		From:      from,
		To:        from,
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       DefaultGasLimit * DefaultGasPrice,
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = hash

	beforeNonce := NodeNonce.Get(from)
	beforeBalance := GetBalance(from)

	err = ApplyTransaction(tx)
	if err == nil {
		t.Fatal("expected transaction to fail")
	}

	if NodeNonce.Get(from) != beforeNonce {
		t.Fatalf(
			"nonce changed after failed transaction: got %d want %d",
			NodeNonce.Get(from),
			beforeNonce,
		)
	}

	if GetBalance(from) != beforeBalance {
		t.Fatalf(
			"balance changed after failed transaction: got %d want %d",
			GetBalance(from),
			beforeBalance,
		)
	}

	if NodeReplay.Exists(tx.Hash) {
		t.Fatal("failed transaction remained in replay state")
	}
}

func TestApplyTransactionRejectsOverflowingTotalWithoutStateChange(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
	}()

	from := "0x1111111111111111111111111111111111111111"
	to := "0x2222222222222222222222222222222222222222"

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: ^uint64(0),
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	tx := Transaction{
		ID:        "atomic-total-overflow",
		From:      from,
		To:        to,
		Amount:    ^uint64(0),
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       1,
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = hash

	beforeBalance := GetBalance(from)
	beforeNonce := NodeNonce.Get(from)

	err = ApplyTransaction(tx)
	if err == nil {
		t.Fatal("expected overflowing transaction to fail")
	}

	if GetBalance(from) != beforeBalance {
		t.Fatalf(
			"sender balance changed after overflow rejection: got %d want %d",
			GetBalance(from),
			beforeBalance,
		)
	}

	if NodeNonce.Get(from) != beforeNonce {
		t.Fatalf(
			"nonce changed after overflow rejection: got %d want %d",
			NodeNonce.Get(from),
			beforeNonce,
		)
	}

	if NodeReplay.Exists(tx.Hash) {
		t.Fatal("rejected overflow transaction remained in replay state")
	}
}

func TestApplyTransactionConcurrentSameNonce(t *testing.T) {
	setupValidFeeEnvironment(t)
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

	NodeNonce = newTestNonceManager()

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	CurrentRewardPool = RewardPool{}
	RewardHistory = nil

	const workers = 128

	transactions := make([]Transaction, workers)

	for i := 0; i < workers; i++ {
		tx := NewTransaction(
			wallet.Address,
			receiver.Address,
			100,
		)

		// Every transaction deliberately uses the same nonce.
		tx.Nonce = 1

		hash, err := GenerateTransactionHash(tx)
		if err != nil {
			t.Fatalf("failed to regenerate transaction hash: %v", err)
		}

		tx.Hash = hash
		transactions[i] = tx
	}

	var wg sync.WaitGroup
	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(tx Transaction) {
			defer wg.Done()
			results <- ApplyTransaction(tx)
		}(transactions[i])
	}

	wg.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful transaction, got %d",
			successes,
		)
	}

	if got := NodeNonce.Get(wallet.Address); got != 1 {
		t.Fatalf(
			"unexpected final nonce: got %d want 1",
			got,
		)
	}

	if got := GetBalance(receiver.Address); got != 100 {
		t.Fatalf(
			"unexpected receiver balance: got %d want 100",
			got,
		)
	}
}

func TestApplyTransactionRewardFailureRollsBackEntireState(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalReplay := NodeReplay
	originalValidators := Validators
	originalRewardPool := CurrentRewardPool
	originalRewardHistory := RewardHistory
	originalTreasury := NetworkTreasury
	originalTreasuryHistory := TreasuryHistory

	defer func() {
		WalletBalances = originalBalances
		NodeNonce = originalNonce
		NodeReplay = originalReplay
		Validators = originalValidators
		CurrentRewardPool = originalRewardPool
		RewardHistory = originalRewardHistory
		NetworkTreasury = originalTreasury
		TreasuryHistory = originalTreasuryHistory
	}()

	const (
		from = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		to   = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)

	Validators = []Validator{
		{
			ID:      1,
			Address: from,
			Power:   100,
			Active:  true,
			Jailed:  false,
			Genesis: true,
		},
	}

	WalletBalances = []WalletBalance{
		{
			Address: from,
			Balance: 100000,
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	// Force reward distribution to fail before it mutates reward state.
	CurrentRewardPool = RewardPool{
		Validator: ^uint64(0),
	}

	CurrentRewardPoolBefore := CurrentRewardPool
	RewardHistory = nil
	NetworkTreasury = Treasury{}
	TreasuryHistory = nil

	tx := Transaction{
		ID:        "reward-failure-rollback",
		From:      from,
		To:        to,
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       10,
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = hash

	beforeSender := GetBalance(from)
	beforeReceiver := GetBalance(to)
	beforeNonce := NodeNonce.Get(from)
	beforeRewardHistory := len(RewardHistory)
	beforeTreasuryHistory := len(TreasuryHistory)
	beforeTreasury := NetworkTreasury

	err = ApplyTransaction(tx)
	if err == nil {
		t.Fatal("expected reward distribution failure")
	}

	if GetBalance(from) != beforeSender {
		t.Fatalf(
			"sender balance was not rolled back: got %d want %d",
			GetBalance(from),
			beforeSender,
		)
	}

	if GetBalance(to) != beforeReceiver {
		t.Fatalf(
			"receiver balance changed after rollback: got %d want %d",
			GetBalance(to),
			beforeReceiver,
		)
	}

	if NodeNonce.Get(from) != beforeNonce {
		t.Fatalf(
			"nonce was not rolled back: got %d want %d",
			NodeNonce.Get(from),
			beforeNonce,
		)
	}

	if NodeReplay.Exists(tx.Hash) {
		t.Fatal("failed transaction remained in replay state")
	}

	if CurrentRewardPool != CurrentRewardPoolBefore {
		t.Fatal("reward pool changed despite failed reward distribution")
	}

	if len(RewardHistory) != beforeRewardHistory {
		t.Fatalf("reward history changed after failed transaction")
	}

	if len(TreasuryHistory) != beforeTreasuryHistory {
		t.Fatalf("treasury history changed after failed transaction")
	}

	if NetworkTreasury != beforeTreasury {
		t.Fatal("treasury changed after failed transaction")
	}
}
