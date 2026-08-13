package app

import (
	"testing"
	"time"
)

func testAdmissionTx(hash, from string, nonce uint64) Transaction {
	return Transaction{
		ID:        hash,
		Hash:      hash,
		From:      from,
		To:        "0x2222222222222222222222222222222222222222",
		Amount:    1,
		Fee:       1,
		Nonce:     nonce,
		Timestamp: time.Now().UTC(),
	}
}

func TestMempoolAdmitRejectsDuplicateHash(t *testing.T) {
	m := NewMempool()

	tx1 := testAdmissionTx("hash-1", "alice", 1)
	tx2 := testAdmissionTx("hash-1", "bob", 2)

	if err := m.AdmitTransaction(tx1); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(tx2); err != ErrMempoolAdmissionDuplicateHash {
		t.Fatalf("expected duplicate hash error, got %v", err)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf("expected 1 transaction, got %d", got)
	}
}

func TestMempoolAdmitRejectsDuplicateSenderNonce(t *testing.T) {
	m := NewMempool()

	tx1 := testAdmissionTx("hash-1", "alice", 7)
	tx2 := testAdmissionTx("hash-2", "alice", 7)

	if err := m.AdmitTransaction(tx1); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(tx2); err != ErrMempoolAdmissionDuplicateNonce {
		t.Fatalf("expected duplicate nonce error, got %v", err)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf("expected 1 transaction, got %d", got)
	}
}

func TestMempoolAdmitAcceptsDifferentNonce(t *testing.T) {
	m := NewMempool()

	tx1 := testAdmissionTx("hash-1", "alice", 1)
	tx2 := testAdmissionTx("hash-2", "alice", 2)

	if err := m.AdmitTransaction(tx1); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(tx2); err != nil {
		t.Fatalf("second admission failed: %v", err)
	}

	if got := m.Count(); got != 2 {
		t.Fatalf("expected 2 transactions, got %d", got)
	}
}

func TestMempoolAdmitAllowsSameNonceForDifferentSenders(t *testing.T) {
	m := NewMempool()

	tx1 := testAdmissionTx("hash-1", "alice", 1)
	tx2 := testAdmissionTx("hash-2", "bob", 1)

	if err := m.AdmitTransaction(tx1); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(tx2); err != nil {
		t.Fatalf("second admission failed: %v", err)
	}

	if got := m.Count(); got != 2 {
		t.Fatalf("expected 2 transactions, got %d", got)
	}
}

func TestMempoolAdmitEnforcesSenderPendingLimit(t *testing.T) {
	m := NewMempool()

	oldLoad := NodeDynamicFee.Load()
	defer func() {
		if err := NodeDynamicFee.SetLoadPercent(oldLoad); err != nil {
			t.Fatal(err)
		}
	}()

	if err := NodeDynamicFee.SetLoadPercent(0); err != nil {
		t.Fatal(err)
	}

	for i := uint64(0); i < NormalSenderPendingLimit; i++ {
		tx := testAdmissionTx(
			"limit-hash-"+string(rune(i+1)),
			"alice",
			i+1,
		)

		if err := m.AdmitTransaction(tx); err != nil {
			t.Fatalf("admission failed at transaction %d: %v", i, err)
		}
	}

	extra := testAdmissionTx("limit-extra", "alice", NormalSenderPendingLimit+1)

	if err := m.AdmitTransaction(extra); err != ErrMempoolAdmissionSenderLimit {
		t.Fatalf("expected sender limit error, got %v", err)
	}

	if got := m.Count(); got != int(NormalSenderPendingLimit) {
		t.Fatalf(
			"expected %d transactions, got %d",
			NormalSenderPendingLimit,
			got,
		)
	}
}
