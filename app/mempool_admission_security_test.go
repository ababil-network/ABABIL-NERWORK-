package app

import (
	"fmt"
	"sync"
	"testing"
)

func TestSenderPendingLimitByLoad(t *testing.T) {
	tests := []struct {
		load uint64
		want uint64
	}{
		{0, 256},
		{50, 256},
		{94, 256},
		{95, 128},
		{96, 102},
		{97, 76},
		{98, 51},
		{99, 25},
		{100, 16},
	}

	for _, tt := range tests {
		if got := SenderPendingLimit(tt.load); got != tt.want {
			t.Fatalf(
				"load=%d: got sender limit %d, want %d",
				tt.load,
				got,
				tt.want,
			)
		}
	}
}

func TestMempoolAdmissionRejectsDuplicateHash(t *testing.T) {
	m := NewMempool()

	tx := Transaction{
		Hash:  "duplicate-hash",
		From:  "sender",
		Nonce: 1,
	}

	if err := m.AdmitTransaction(tx); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(tx); err != ErrMempoolAdmissionDuplicateHash {
		t.Fatalf(
			"expected duplicate hash error, got %v",
			err,
		)
	}

	if m.Count() != 1 {
		t.Fatalf("expected exactly one transaction, got %d", m.Count())
	}
}

func TestMempoolAdmissionRejectsDuplicateSenderNonce(t *testing.T) {
	m := NewMempool()

	first := Transaction{
		Hash:  "hash-1",
		From:  "sender",
		Nonce: 7,
	}

	second := Transaction{
		Hash:  "hash-2",
		From:  "sender",
		Nonce: 7,
	}

	if err := m.AdmitTransaction(first); err != nil {
		t.Fatalf("first admission failed: %v", err)
	}

	if err := m.AdmitTransaction(second); err != ErrMempoolAdmissionDuplicateNonce {
		t.Fatalf(
			"expected duplicate nonce error, got %v",
			err,
		)
	}

	if m.Count() != 1 {
		t.Fatalf("expected exactly one transaction, got %d", m.Count())
	}
}

func TestMempoolAdmissionEnforcesSenderLimit(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()

	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	for i := uint64(0); i < Congestion100SenderPendingLimit; i++ {
		tx := Transaction{
			Hash:  fmt.Sprintf("hash-%d", i),
			From:  "sender",
			Nonce: i + 1,
		}

		if err := m.AdmitTransaction(tx); err != nil {
			t.Fatalf(
				"transaction %d should have been admitted: %v",
				i,
				err,
			)
		}
	}

	extra := Transaction{
		Hash:  "hash-extra",
		From:  "sender",
		Nonce: Congestion100SenderPendingLimit + 1,
	}

	if err := m.AdmitTransaction(extra); err != ErrMempoolAdmissionSenderLimit {
		t.Fatalf(
			"expected sender limit error, got %v",
			err,
		)
	}

	if got := m.Count(); got != int(Congestion100SenderPendingLimit) {
		t.Fatalf(
			"unexpected mempool count: got %d want %d",
			got,
			Congestion100SenderPendingLimit,
		)
	}
}

func TestMempoolAdmissionConcurrentSenderLimit(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()

	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	const workers = 128

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			tx := Transaction{
				Hash:  fmt.Sprintf("concurrent-hash-%d", i),
				From:  "same-sender",
				Nonce: uint64(i + 1),
			}

			results <- m.AdmitTransaction(tx)
		}()
	}

	wg.Wait()
	close(results)

	successes := 0

	for err := range results {
		if err == nil {
			successes++
			continue
		}

		if err != ErrMempoolAdmissionSenderLimit {
			t.Fatalf(
				"unexpected admission error: %v",
				err,
			)
		}
	}

	if successes != int(Congestion100SenderPendingLimit) {
		t.Fatalf(
			"expected %d successful admissions, got %d",
			Congestion100SenderPendingLimit,
			successes,
		)
	}

	if got := m.Count(); got != int(Congestion100SenderPendingLimit) {
		t.Fatalf(
			"concurrent admission exceeded sender limit: got %d want %d",
			got,
			Congestion100SenderPendingLimit,
		)
	}
}
