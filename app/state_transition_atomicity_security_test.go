package app

import (
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
