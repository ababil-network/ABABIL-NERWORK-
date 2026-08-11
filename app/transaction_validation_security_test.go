package app

import (
	"math"
	"testing"
	"time"
)

func setupValidationSecurityTest(t *testing.T) (*Wallet, *Wallet) {
	t.Helper()

	sender, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}

	receiver, err := CreateWallet()
	if err != nil {
		t.Fatalf("failed to create receiver: %v", err)
	}

	NodeWallet = sender

	WalletBalances = []WalletBalance{
		{
			Address: sender.Address,
			Balance: math.MaxUint64,
		},
	}

	NodeNonce = &NonceManager{
		nonces: make(map[string]uint64),
	}

	NodeFreeTransaction = &FreeTransactionManager{
		data: make(map[string]*FreeTransactionInfo),
	}

	NodeReplay = &ReplayManager{
		seen: make(map[string]bool),
	}

	// Deterministic reference price for fee validation tests.
	// 1 ABABIL = $0.10.
	NodeReferencePrice.Reset()
	if err := NodeReferencePrice.AddObservation(
		100_000,
		time.Now().UTC(),
	); err != nil {
		t.Fatalf("failed to initialize reference price: %v", err)
	}

	// Start security tests from the lowest-load fee tier.
	if err := NodeDynamicFee.SetLoadPercent(0); err != nil {
		t.Fatalf("failed to initialize fee load: %v", err)
	}

	return sender, receiver
}

func restoreValidationSecurityState(
	balances []WalletBalance,
	nonce *NonceManager,
	free *FreeTransactionManager,
	replay *ReplayManager,
	wallet *Wallet,
) {
	WalletBalances = balances
	NodeNonce = nonce
	NodeFreeTransaction = free
	NodeReplay = replay
	NodeWallet = wallet
}

func validSecurityTransaction(t *testing.T) (Transaction, *Wallet, *Wallet) {
	t.Helper()

	sender, receiver := setupValidationSecurityTest(t)

	tx := NewTransaction(
		sender.Address,
		receiver.Address,
		100,
	)

	if tx.Hash == "" || tx.Signature == "" || tx.PublicKey == "" {
		t.Fatal("NewTransaction did not produce a complete signed transaction")
	}

	return tx, sender, receiver
}

func TestValidateTransactionRejectsZeroAmount(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Amount = 0

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("zero-amount transaction was accepted")
	}
}

func TestValidateTransactionRejectsZeroGasLimit(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.GasLimit = 0

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("zero gas-limit transaction was accepted")
	}
}

func TestValidateTransactionRejectsExcessiveGasLimit(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.GasLimit = MaxGasLimit + 1

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("excessive gas-limit transaction was accepted")
	}
}

func TestValidateTransactionRejectsSameSenderAndReceiver(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, sender, _ := validSecurityTransaction(t)

	tx.To = sender.Address

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("self-transfer transaction was accepted")
	}
}

func TestValidateTransactionRejectsTamperedHash(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Amount++

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("transaction with tampered amount and original hash was accepted")
	}
}

func TestValidateTransactionRejectsTamperedSignature(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	// Deterministically tamper with the recovery byte.
	// Ethereum-style signatures use V = 0 or 1, so flip it
	// instead of forcing it to 0 (which could leave an existing
	// V=0 signature unchanged).
	last := tx.Signature[len(tx.Signature)-2:]
	if last == "00" {
		tx.Signature = tx.Signature[:len(tx.Signature)-2] + "01"
	} else {
		tx.Signature = tx.Signature[:len(tx.Signature)-2] + "00"
	}

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("tampered signature was accepted")
	}
}

func TestValidateTransactionRejectsWrongNonce(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Nonce++

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	signed := SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("wrong nonce was accepted")
	}
}

func TestValidateTransactionRejectsIncorrectFee(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Fee++

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	signed := SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("incorrect transaction fee was accepted")
	}
}

func TestValidateTransactionRejectsIncorrectFinalNativeFee(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	// This wallet still has free-transaction quota, so explicitly
	// exhaust it before testing the paid final-fee path.
	for NodeFreeTransaction.Remaining(tx.From) > 0 {
		if !NodeFreeTransaction.Use(tx.From) {
			t.Fatal("failed to consume free transaction quota")
		}
	}

	expectedFee, err := CalculateFinalNativeFee()
	if err != nil {
		t.Fatalf("failed to calculate final native fee: %v", err)
	}

	if expectedFee == 0 {
		t.Fatal("final native fee must be greater than zero")
	}

	tx.Fee = expectedFee + 1
	tx.GasPrice = 0

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = hash

	signed := SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("incorrect final native transaction fee was accepted")
	}
}

func TestValidateTransactionRejectsAmountFeeOverflow(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Amount = math.MaxUint64
	tx.Fee = 1

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	signed := SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	if err := ValidateTransaction(tx); err != ErrTransactionValueOverflow {
		t.Fatalf(
			"expected ErrTransactionValueOverflow, got %v",
			err,
		)
	}
}

func TestValidateTransactionRejectsInsufficientBalance(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, sender, _ := validSecurityTransaction(t)

	WalletBalances = []WalletBalance{
		{
			Address: sender.Address,
			Balance: 1,
		},
	}

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("insufficient-balance transaction was accepted")
	}
}

func TestValidateTransactionRejectsZeroTimestamp(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, _, _ := validSecurityTransaction(t)

	tx.Timestamp = time.Time{}

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("zero timestamp transaction was accepted")
	}
}

func TestValidateTransactionRejectsExhaustedFreeQuota(t *testing.T) {
	originalBalances := WalletBalances
	originalNonce := NodeNonce
	originalFree := NodeFreeTransaction
	originalReplay := NodeReplay
	originalWallet := NodeWallet

	defer restoreValidationSecurityState(
		originalBalances,
		originalNonce,
		originalFree,
		originalReplay,
		originalWallet,
	)

	tx, sender, _ := validSecurityTransaction(t)

	// Consume the complete free quota.
	for NodeFreeTransaction.Remaining(sender.Address) > 0 {
		if !NodeFreeTransaction.Use(sender.Address) {
			t.Fatal("failed to consume free transaction quota")
		}
	}

	tx.Fee = 0

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}
	tx.Hash = hash

	signed := SignTransactionWithPrivateKey(hash, NodeWallet.PrivateKey)
	tx.Signature = signed.Signature
	tx.PublicKey = signed.PublicKey

	if err := ValidateTransaction(tx); err == nil {
		t.Fatal("zero-fee transaction bypassed exhausted free quota")
	}
}
