package app

import (
	"testing"
	"time"
)

func TestMempoolPriorityHigherFeeFirst(t *testing.T) {
	mempool := NewMempool()

	now := time.Now().UTC()

	low := Transaction{
		Hash:      "low",
		Fee:       10,
		Timestamp: now.Add(2 * time.Second),
	}

	high := Transaction{
		Hash:      "high",
		Fee:       20,
		Timestamp: now.Add(3 * time.Second),
	}

	mempool.AddTransaction(low)
	mempool.AddTransaction(high)

	snapshot := mempool.Snapshot()

	if len(snapshot) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(snapshot))
	}

	if snapshot[0].Hash != "high" {
		t.Fatalf(
			"higher-fee transaction was not prioritized: first=%s",
			snapshot[0].Hash,
		)
	}
}

func TestMempoolPriorityEqualFeeOldestFirst(t *testing.T) {
	mempool := NewMempool()

	now := time.Now().UTC()

	oldest := Transaction{
		Hash:      "oldest",
		Fee:       100,
		Timestamp: now,
	}

	newer := Transaction{
		Hash:      "newer",
		Fee:       100,
		Timestamp: now.Add(time.Second),
	}

	mempool.AddTransaction(newer)
	mempool.AddTransaction(oldest)

	snapshot := mempool.Snapshot()

	if len(snapshot) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(snapshot))
	}

	if snapshot[0].Hash != "oldest" {
		t.Fatalf(
			"equal-fee ordering is not oldest-first: first=%s",
			snapshot[0].Hash,
		)
	}
}

func TestMempoolPriorityDeterministicHashTieBreak(t *testing.T) {
	mempool := NewMempool()

	timestamp := time.Unix(1000, 0).UTC()

	txA := Transaction{
		Hash:      "aaa",
		Fee:       100,
		Timestamp: timestamp,
	}

	txB := Transaction{
		Hash:      "bbb",
		Fee:       100,
		Timestamp: timestamp,
	}

	mempool.AddTransaction(txB)
	mempool.AddTransaction(txA)

	snapshot := mempool.Snapshot()

	if len(snapshot) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(snapshot))
	}

	if snapshot[0].Hash != "aaa" {
		t.Fatalf(
			"equal-fee/equal-time ordering is not deterministic: first=%s",
			snapshot[0].Hash,
		)
	}
}

func TestMempoolPriorityNeverMutatesTransactionData(t *testing.T) {
	mempool := NewMempool()

	tx := Transaction{
		Hash:      "immutable",
		Fee:       123,
		Amount:    456,
		Nonce:     7,
		Timestamp: time.Now().UTC(),
	}

	mempool.AddTransaction(tx)

	snapshot := mempool.Snapshot()

	if len(snapshot) != 1 {
		t.Fatalf("expected one transaction, got %d", len(snapshot))
	}

	got := snapshot[0]

	if got.Hash != tx.Hash ||
		got.Fee != tx.Fee ||
		got.Amount != tx.Amount ||
		got.Nonce != tx.Nonce {
		t.Fatal("mempool priority mutated transaction data")
	}
}
