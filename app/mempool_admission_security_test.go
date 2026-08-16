package app

import (
	"fmt"
	"sync"
	"testing"
	"time"
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

func TestMempoolAdmitTransactionsAtomicDuplicateHash(t *testing.T) {
	m := NewMempool()

	existing := Transaction{
		Hash:  "existing-hash",
		From:  "alice",
		Nonce: 1,
	}

	if err := m.AdmitTransaction(existing); err != nil {
		t.Fatalf("existing admission failed: %v", err)
	}

	batch := []Transaction{
		{
			Hash:  "batch-1",
			From:  "bob",
			Nonce: 1,
		},
		existing,
		{
			Hash:  "batch-3",
			From:  "charlie",
			Nonce: 1,
		},
	}

	if err := m.AdmitTransactions(batch); err != ErrMempoolAdmissionDuplicateHash {
		t.Fatalf("expected duplicate hash error, got %v", err)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf("batch was not atomic: got %d transactions, want 1", got)
	}

	if _, exists := m.hashes["batch-1"]; exists {
		t.Fatal("batch transaction was partially committed")
	}

	if _, exists := m.hashes["batch-3"]; exists {
		t.Fatal("batch transaction was partially committed")
	}
}

func TestMempoolAdmitTransactionsAtomicDuplicateNonce(t *testing.T) {
	m := NewMempool()

	batch := []Transaction{
		{
			Hash:  "nonce-1",
			From:  "alice",
			Nonce: 10,
		},
		{
			Hash:  "nonce-2",
			From:  "alice",
			Nonce: 10,
		},
		{
			Hash:  "nonce-3",
			From:  "bob",
			Nonce: 1,
		},
	}

	if err := m.AdmitTransactions(batch); err != ErrMempoolAdmissionDuplicateNonce {
		t.Fatalf("expected duplicate nonce error, got %v", err)
	}

	if got := m.Count(); got != 0 {
		t.Fatalf("batch was not atomic: got %d transactions, want 0", got)
	}

	if len(m.hashes) != 0 ||
		len(m.senderNonces) != 0 ||
		len(m.senderCounts) != 0 {
		t.Fatal("mempool indexes changed after rejected batch")
	}
}

func TestMempoolAdmitTransactionsAtomicExistingNonce(t *testing.T) {
	m := NewMempool()

	existing := Transaction{
		Hash:  "existing-nonce",
		From:  "alice",
		Nonce: 7,
	}

	if err := m.AdmitTransaction(existing); err != nil {
		t.Fatalf("existing admission failed: %v", err)
	}

	batch := []Transaction{
		{
			Hash:  "new-1",
			From:  "bob",
			Nonce: 1,
		},
		{
			Hash:  "new-2",
			From:  "alice",
			Nonce: 7,
		},
		{
			Hash:  "new-3",
			From:  "charlie",
			Nonce: 1,
		},
	}

	if err := m.AdmitTransactions(batch); err != ErrMempoolAdmissionDuplicateNonce {
		t.Fatalf("expected duplicate nonce error, got %v", err)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf("batch was not atomic: got %d transactions, want 1", got)
	}

	if _, exists := m.hashes["new-1"]; exists {
		t.Fatal("new-1 was partially committed")
	}

	if _, exists := m.hashes["new-3"]; exists {
		t.Fatal("new-3 was partially committed")
	}
}

func TestMempoolAdmitTransactionsAtomicSenderLimit(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()
	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	for i := uint64(0); i < Congestion100SenderPendingLimit-1; i++ {
		tx := Transaction{
			Hash:  fmt.Sprintf("existing-%d", i),
			From:  "alice",
			Nonce: i + 1,
		}

		if err := m.AdmitTransaction(tx); err != nil {
			t.Fatalf("existing admission %d failed: %v", i, err)
		}
	}

	batch := []Transaction{
		{
			Hash:  "limit-1",
			From:  "alice",
			Nonce: 1000,
		},
		{
			Hash:  "limit-2",
			From:  "alice",
			Nonce: 1001,
		},
		{
			Hash:  "other-sender",
			From:  "bob",
			Nonce: 1,
		},
	}

	if err := m.AdmitTransactions(batch); err != ErrMempoolAdmissionSenderLimit {
		t.Fatalf("expected sender limit error, got %v", err)
	}

	if got := m.Count(); got != int(Congestion100SenderPendingLimit-1) {
		t.Fatalf(
			"batch was not atomic: got %d transactions, want %d",
			got,
			Congestion100SenderPendingLimit-1,
		)
	}

	if _, exists := m.hashes["limit-1"]; exists {
		t.Fatal("limit-1 was partially committed")
	}

	if _, exists := m.hashes["limit-2"]; exists {
		t.Fatal("limit-2 was partially committed")
	}

	if _, exists := m.hashes["other-sender"]; exists {
		t.Fatal("other-sender was partially committed")
	}
}

func TestMempoolAdmitTransactionsAtomicCapacity(t *testing.T) {
	m := NewMempool()

	for i := 0; i < MaxMempoolTransactions-1; i++ {
		tx := Transaction{
			Hash:  fmt.Sprintf("capacity-existing-%d", i),
			From:  fmt.Sprintf("sender-%d", i),
			Nonce: 1,
		}

		m.AddTransaction(tx)
	}

	batch := []Transaction{
		{
			Hash:  "capacity-new-1",
			From:  "capacity-sender-1",
			Nonce: 1,
		},
		{
			Hash:  "capacity-new-2",
			From:  "capacity-sender-2",
			Nonce: 1,
		},
	}

	if err := m.AdmitTransactions(batch); err != ErrMempoolAdmissionFull {
		t.Fatalf("expected mempool full error, got %v", err)
	}

	if got := m.Count(); got != MaxMempoolTransactions-1 {
		t.Fatalf(
			"capacity rejection was not atomic: got %d transactions, want %d",
			got,
			MaxMempoolTransactions-1,
		)
	}

	if _, exists := m.hashes["capacity-new-1"]; exists {
		t.Fatal("capacity-new-1 was partially committed")
	}

	if _, exists := m.hashes["capacity-new-2"]; exists {
		t.Fatal("capacity-new-2 was partially committed")
	}
}

func TestMempoolAdmitTransactionsMixedSenderLimits(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()
	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	batch := []Transaction{
		{Hash: "alice-1", From: "alice", Nonce: 1},
		{Hash: "alice-2", From: "alice", Nonce: 2},
		{Hash: "bob-1", From: "bob", Nonce: 1},
		{Hash: "bob-2", From: "bob", Nonce: 2},
	}

	if err := m.AdmitTransactions(batch); err != nil {
		t.Fatalf("mixed-sender batch admission failed: %v", err)
	}

	if got := m.Count(); got != len(batch) {
		t.Fatalf("got %d transactions, want %d", got, len(batch))
	}

	if got := m.senderCounts["alice"]; got != 2 {
		t.Fatalf("alice sender count = %d, want 2", got)
	}

	if got := m.senderCounts["bob"]; got != 2 {
		t.Fatalf("bob sender count = %d, want 2", got)
	}
}

func TestMempoolAdmitTransactionsIndexesRemainConsistentAfterSuccess(t *testing.T) {
	m := NewMempool()

	batch := []Transaction{
		{Hash: "index-1", From: "alice", Nonce: 1},
		{Hash: "index-2", From: "alice", Nonce: 2},
		{Hash: "index-3", From: "bob", Nonce: 1},
	}

	if err := m.AdmitTransactions(batch); err != nil {
		t.Fatalf("batch admission failed: %v", err)
	}

	if len(m.Transactions) != len(m.hashes) {
		t.Fatalf(
			"hash index mismatch: transactions=%d hashes=%d",
			len(m.Transactions),
			len(m.hashes),
		)
	}

	if len(m.Transactions) != len(m.senderNonces) {
		t.Fatalf(
			"sender nonce index mismatch: transactions=%d senderNonces=%d",
			len(m.Transactions),
			len(m.senderNonces),
		)
	}

	if m.senderCounts["alice"] != 2 {
		t.Fatalf("alice count = %d, want 2", m.senderCounts["alice"])
	}

	if m.senderCounts["bob"] != 1 {
		t.Fatalf("bob count = %d, want 1", m.senderCounts["bob"])
	}
}

func TestMempoolAdmitTransactionsConcurrentSenderLimit(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()
	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	const workers = 64
	const batchSize = 2

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			batch := []Transaction{
				{
					Hash:  fmt.Sprintf("batch-%d-tx-1", i),
					From:  "same-sender",
					Nonce: uint64(i*batchSize + 1),
				},
				{
					Hash:  fmt.Sprintf("batch-%d-tx-2", i),
					From:  "same-sender",
					Nonce: uint64(i*batchSize + 2),
				},
			}

			results <- m.AdmitTransactions(batch)
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
			t.Fatalf("unexpected concurrent batch error: %v", err)
		}
	}

	want := int(Congestion100SenderPendingLimit)

	if got := m.Count(); got != want {
		t.Fatalf(
			"concurrent batches exceeded sender limit: got %d want %d",
			got,
			want,
		)
	}

	if m.senderCounts["same-sender"] != Congestion100SenderPendingLimit {
		t.Fatalf(
			"sender count mismatch: got %d want %d",
			m.senderCounts["same-sender"],
			Congestion100SenderPendingLimit,
		)
	}

	if successes != want/batchSize {
		t.Fatalf(
			"unexpected successful batch count: got %d want %d",
			successes,
			want/batchSize,
		)
	}

	if len(m.Transactions) != len(m.hashes) ||
		len(m.Transactions) != len(m.senderNonces) {
		t.Fatalf(
			"index mismatch after concurrent admission: tx=%d hashes=%d nonces=%d",
			len(m.Transactions),
			len(m.hashes),
			len(m.senderNonces),
		)
	}
}

func TestMempoolAdmitTransactionsConcurrentDuplicateNonce(t *testing.T) {
	m := NewMempool()

	const workers = 64

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			batch := []Transaction{
				{
					Hash:  fmt.Sprintf("duplicate-nonce-%d-a", i),
					From:  "same-sender",
					Nonce: 777,
				},
			}

			results <- m.AdmitTransactions(batch)
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

		if err != ErrMempoolAdmissionDuplicateNonce {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful admission, got %d",
			successes,
		)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf(
			"duplicate nonce admitted concurrently: got %d transactions",
			got,
		)
	}

	if got := m.senderCounts["same-sender"]; got != 1 {
		t.Fatalf(
			"sender count mismatch: got %d want 1",
			got,
		)
	}

	if len(m.hashes) != 1 || len(m.senderNonces) != 1 {
		t.Fatalf(
			"index mismatch: hashes=%d nonces=%d",
			len(m.hashes),
			len(m.senderNonces),
		)
	}
}

func TestMempoolAdmitTransactionsConcurrentDuplicateHash(t *testing.T) {
	m := NewMempool()

	const workers = 64

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			batch := []Transaction{
				{
					Hash:  "shared-hash",
					From:  fmt.Sprintf("sender-%d", i),
					Nonce: uint64(i + 1),
				},
			}

			results <- m.AdmitTransactions(batch)
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

		if err != ErrMempoolAdmissionDuplicateHash {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful admission, got %d",
			successes,
		)
	}

	if got := m.Count(); got != 1 {
		t.Fatalf(
			"duplicate hash admitted concurrently: got %d transactions",
			got,
		)
	}

	if len(m.hashes) != 1 ||
		len(m.senderNonces) != 1 ||
		len(m.senderCounts) != 1 {
		t.Fatalf(
			"index mismatch: hashes=%d nonces=%d senders=%d",
			len(m.hashes),
			len(m.senderNonces),
			len(m.senderCounts),
		)
	}
}

func TestMempoolAdmitTransactionsConcurrentMixedSenders(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()
	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(100); err != nil {
		t.Fatalf("failed to set load: %v", err)
	}

	m := NewMempool()

	const workers = 32
	const batchSize = 2

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			sender := fmt.Sprintf("sender-%d", i%4)

			batch := []Transaction{
				{
					Hash:  fmt.Sprintf("mixed-%d-a", i),
					From:  sender,
					Nonce: uint64(i*batchSize + 1),
				},
				{
					Hash:  fmt.Sprintf("mixed-%d-b", i),
					From:  sender,
					Nonce: uint64(i*batchSize + 2),
				},
			}

			results <- m.AdmitTransactions(batch)
		}()
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil &&
			err != ErrMempoolAdmissionSenderLimit {
			t.Fatalf("unexpected concurrent mixed-sender error: %v", err)
		}
	}

	total := 0

	for sender, count := range m.senderCounts {
		if count > Congestion100SenderPendingLimit {
			t.Fatalf(
				"sender limit exceeded for %s: got %d want <= %d",
				sender,
				count,
				Congestion100SenderPendingLimit,
			)
		}

		total += int(count)
	}

	if total != m.Count() {
		t.Fatalf(
			"sender count/index mismatch: senderTotal=%d mempoolCount=%d",
			total,
			m.Count(),
		)
	}

	if len(m.Transactions) != len(m.hashes) ||
		len(m.Transactions) != len(m.senderNonces) {
		t.Fatalf(
			"index mismatch: tx=%d hashes=%d nonces=%d",
			len(m.Transactions),
			len(m.hashes),
			len(m.senderNonces),
		)
	}
}

func TestMempoolAdmitTransactionsConcurrentRejectedBatchLeavesNoPartialState(t *testing.T) {
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
			Hash:  fmt.Sprintf("preloaded-%d", i),
			From:  "protected-sender",
			Nonce: i + 1,
		}

		if err := m.AdmitTransaction(tx); err != nil {
			t.Fatalf("preload admission %d failed: %v", i, err)
		}
	}

	const workers = 32

	var wg sync.WaitGroup
	wg.Add(workers)

	results := make(chan error, workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			batch := []Transaction{
				{
					Hash:  fmt.Sprintf("rejected-%d-a", i),
					From:  "protected-sender",
					Nonce: uint64(1000 + i*2),
				},
				{
					Hash:  fmt.Sprintf("rejected-%d-b", i),
					From:  "protected-sender",
					Nonce: uint64(2000 + i*2),
				},
			}

			results <- m.AdmitTransactions(batch)
		}()
	}

	wg.Wait()
	close(results)

	for err := range results {
		if err != ErrMempoolAdmissionSenderLimit {
			t.Fatalf("unexpected rejection error: %v", err)
		}
	}

	if got := m.Count(); got != int(Congestion100SenderPendingLimit) {
		t.Fatalf(
			"rejected concurrent batches changed mempool: got %d want %d",
			got,
			Congestion100SenderPendingLimit,
		)
	}

	if got := m.senderCounts["protected-sender"]; got != Congestion100SenderPendingLimit {
		t.Fatalf(
			"rejected concurrent batches changed sender count: got %d want %d",
			got,
			Congestion100SenderPendingLimit,
		)
	}

	if len(m.hashes) != int(Congestion100SenderPendingLimit) ||
		len(m.senderNonces) != int(Congestion100SenderPendingLimit) {
		t.Fatalf(
			"rejected concurrent batches changed indexes: hashes=%d nonces=%d",
			len(m.hashes),
			len(m.senderNonces),
		)
	}
}

func TestMempoolIndexesRemainConsistentAfterBatchRemoval(t *testing.T) {
	m := NewMempool()

	txs := make([]Transaction, 32)

	for i := range txs {
		txs[i] = Transaction{
			Hash:  fmt.Sprintf("index-consistency-%d", i),
			From:  fmt.Sprintf("sender-%d", i%4),
			Nonce: uint64(i + 1),
		}
	}

	if err := m.AdmitTransactions(txs); err != nil {
		t.Fatalf("batch admission failed: %v", err)
	}

	if len(m.Transactions) != 32 ||
		len(m.hashes) != 32 ||
		len(m.senderNonces) != 32 {
		t.Fatalf(
			"indexes inconsistent after admission: tx=%d hashes=%d nonces=%d",
			len(m.Transactions),
			len(m.hashes),
			len(m.senderNonces),
		)
	}

	for sender := 0; sender < 4; sender++ {
		key := fmt.Sprintf("sender-%d", sender)

		if got := m.senderCounts[key]; got != 8 {
			t.Fatalf(
				"sender count mismatch before removal for %s: got %d want 8",
				key,
				got,
			)
		}
	}

	processed := txs[:16]

	m.RemoveProcessedTransactions(processed)

	if m.Count() != 16 {
		t.Fatalf("expected 16 remaining transactions, got %d", m.Count())
	}

	if len(m.hashes) != 16 ||
		len(m.senderNonces) != 16 {
		t.Fatalf(
			"indexes inconsistent after removal: hashes=%d nonces=%d",
			len(m.hashes),
			len(m.senderNonces),
		)
	}

	totalSenderCount := uint64(0)

	for sender, count := range m.senderCounts {
		if count == 0 {
			t.Fatalf("zero sender count retained for %s", sender)
		}

		totalSenderCount += count
	}

	if totalSenderCount != 16 {
		t.Fatalf(
			"sender count total mismatch: got %d want 16",
			totalSenderCount,
		)
	}

	for _, tx := range processed {
		if _, exists := m.hashes[tx.Hash]; exists {
			t.Fatalf("processed hash still indexed: %s", tx.Hash)
		}

		key := mempoolSenderNonceKey{
			From:  tx.From,
			Nonce: tx.Nonce,
		}

		if _, exists := m.senderNonces[key]; exists {
			t.Fatalf(
				"processed sender+nonce still indexed: %s/%d",
				tx.From,
				tx.Nonce,
			)
		}
	}
}

func TestMempoolIndexesRemainConsistentAfterExpiration(t *testing.T) {
	m := NewMempool()

	oldTxs := make([]Transaction, 8)

	for i := range oldTxs {
		oldTxs[i] = Transaction{
			Hash:      fmt.Sprintf("expired-%d", i),
			From:      fmt.Sprintf("expired-sender-%d", i%2),
			Nonce:     uint64(i + 1),
			Timestamp: time.Now().UTC().Add(-MempoolTransactionTTL - time.Second),
		}

		if err := m.AdmitTransaction(oldTxs[i]); err != nil {
			t.Fatalf("expired transaction admission failed: %v", err)
		}
	}

	freshTx := Transaction{
		Hash:      "fresh-transaction",
		From:      "fresh-sender",
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	if err := m.AdmitTransaction(freshTx); err != nil {
		t.Fatalf("fresh transaction admission failed: %v", err)
	}

	m.RemoveExpiredTransactions()

	if m.Count() != 1 {
		t.Fatalf(
			"expected only fresh transaction to remain, got %d",
			m.Count(),
		)
	}

	if len(m.hashes) != 1 ||
		len(m.senderNonces) != 1 ||
		len(m.senderCounts) != 1 {
		t.Fatalf(
			"indexes inconsistent after expiration: tx=%d hashes=%d nonces=%d senders=%d",
			len(m.Transactions),
			len(m.hashes),
			len(m.senderNonces),
			len(m.senderCounts),
		)
	}

	if _, exists := m.hashes[freshTx.Hash]; !exists {
		t.Fatal("fresh transaction hash was removed incorrectly")
	}

	if got := m.senderCounts[freshTx.From]; got != 1 {
		t.Fatalf(
			"fresh sender count incorrect: got %d want 1",
			got,
		)
	}

	for _, tx := range oldTxs {
		if _, exists := m.hashes[tx.Hash]; exists {
			t.Fatalf("expired transaction hash still indexed: %s", tx.Hash)
		}
	}
}

func TestMempoolAdmitTransactionsAdaptiveBatchIndexBoundary(t *testing.T) {
	tests := []struct {
		name      string
		batchSize int
		want16    bool
	}{
		{name: "single", batchSize: 1, want16: true},
		{name: "normal_batch", batchSize: 25000, want16: true},
		{name: "current_mempool_capacity", batchSize: 50000, want16: true},
		{name: "uint16_boundary_minus_one", batchSize: 65534, want16: true},
		{name: "uint16_boundary", batchSize: 65535, want16: true},
		{name: "first_uint32", batchSize: 65536, want16: false},
		{name: "future_100k", batchSize: 100000, want16: false},
		{name: "future_1m", batchSize: 1000000, want16: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := useBatchIndex16(tt.batchSize); got != tt.want16 {
				t.Fatalf(
					"useBatchIndex16(%d) = %v, want %v",
					tt.batchSize,
					got,
					tt.want16,
				)
			}
		})
	}

	if maxBatchIndex16Entries != int(^uint16(0)) {
		t.Fatalf(
			"unexpected uint16 boundary: got %d want %d",
			maxBatchIndex16Entries,
			int(^uint16(0)),
		)
	}

	if maxBatchIndex16Entries != 65535 {
		t.Fatalf(
			"unexpected uint16 max entry count: got %d want 65535",
			maxBatchIndex16Entries,
		)
	}
}

func TestMempoolAdaptiveBatchIndexLazyUint32Allocation(t *testing.T) {
	oldLoad := NodeDynamicFee.Load()
	defer func() {
		_ = NodeDynamicFee.SetLoadPercent(oldLoad)
	}()

	if err := NodeDynamicFee.SetLoadPercent(0); err != nil {
		t.Fatalf("failed to set deterministic test load: %v", err)
	}

	m := NewMempool()

	if m.batchIndexes16 != nil {
		t.Fatal("uint16 scratch buffer must start unallocated")
	}
	if m.batchIndexes32 != nil {
		t.Fatal("uint32 scratch buffer must start unallocated")
	}

	// 25k entries must take the compact uint16 path.
	// The sender sequence is deliberately unsorted.
	txs := make([]Transaction, 25000)

	const senderCount = 125
	const txsPerSender = 200

	for i := range txs {
		sender := i % senderCount
		nonce := uint64(i/senderCount + 1)

		txs[i] = benchmarkMempoolTransaction(i)
		txs[i].From = fmt.Sprintf("0x%040x", sender+1)
		txs[i].Nonce = nonce
	}

	if len(txs) != senderCount*txsPerSender {
		t.Fatalf("unexpected test batch size: got %d", len(txs))
	}

	// Confirm the input is genuinely unsorted.
	// Confirm the input is genuinely unsorted.
	unsorted := false
	for i := 1; i < len(txs); i++ {
		prev := txs[i-1]
		curr := txs[i]
		if prev.From > curr.From ||
			(prev.From == curr.From && prev.Nonce > curr.Nonce) {
			unsorted = true
			break
		}
	}
	if !unsorted {
		t.Fatal("test batch is unexpectedly sorted")
	}

	if err := m.AdmitTransactions(txs); err != nil {
		t.Fatalf("unexpected admission error: %v", err)
	}

	if m.batchIndexes16 == nil {
		t.Fatal("uint16 scratch buffer was not allocated for a 25k unsorted batch")
	}

	if len(m.batchIndexes16) != len(txs) {
		t.Fatalf(
			"unexpected uint16 scratch length: got %d want %d",
			len(m.batchIndexes16),
			len(txs),
		)
	}

	if m.batchIndexes32 != nil {
		t.Fatal("uint32 scratch buffer was allocated for a 25k batch")
	}
}

func TestMempoolAdaptiveBatchIndexPathSelection(t *testing.T) {
	if !useBatchIndex16(50000) {
		t.Fatal("50,000-entry batch must use uint16 index path")
	}

	if !useBatchIndex16(65535) {
		t.Fatal("65,535-entry batch must use uint16 index path")
	}

	if useBatchIndex16(65536) {
		t.Fatal("65,536-entry batch must use uint32 index path")
	}

	if useBatchIndex16(maxBatchIndex16Entries + 1) {
		t.Fatal("batch above uint16 boundary must use uint32 index path")
	}
}

func TestMempoolRemoveProcessedConcurrent(t *testing.T) {
	m := NewMempool()

	const total = 1000
	txs := make([]Transaction, total)

	for i := range txs {
		txs[i] = Transaction{
			Hash:  fmt.Sprintf("opt17g-remove-%d", i),
			From:  fmt.Sprintf("opt17g-sender-%d", i),
			Nonce: uint64(i + 1),
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("admit tx %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		start := i * 100
		end := start + 100

		wg.Add(1)
		go func(batch []Transaction) {
			defer wg.Done()
			m.RemoveProcessedTransactions(batch)
		}(txs[start:end])
	}

	wg.Wait()

	if got := m.Count(); got != 0 {
		t.Fatalf("concurrent processed removal left %d transactions", got)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.hashes) != 0 {
		t.Fatalf("hash index not empty after concurrent removal: %d", len(m.hashes))
	}

	if len(m.senderNonces) != 0 {
		t.Fatalf("sender nonce index not empty after concurrent removal: %d", len(m.senderNonces))
	}

	if len(m.senderCounts) != 0 {
		t.Fatalf("sender count index not empty after concurrent removal: %d", len(m.senderCounts))
	}
}

func TestMempoolRemoveProcessedConcurrentOverlappingBatches(t *testing.T) {
	m := NewMempool()

	const total = 1000
	txs := make([]Transaction, total)

	for i := range txs {
		txs[i] = Transaction{
			Hash:  fmt.Sprintf("opt17g-overlap-%d", i),
			From:  fmt.Sprintf("opt17g-overlap-sender-%d", i),
			Nonce: uint64(i + 1),
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("admit tx %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		start := (i * 50) % total
		batch := make([]Transaction, 100)

		for j := range batch {
			batch[j] = txs[(start+j)%total]
		}

		wg.Add(1)
		go func(batch []Transaction) {
			defer wg.Done()
			m.RemoveProcessedTransactions(batch)
		}(batch)
	}

	wg.Wait()

	if got := m.Count(); got != 0 {
		t.Fatalf("overlapping concurrent removal left %d transactions", got)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.hashes) != 0 ||
		len(m.senderNonces) != 0 ||
		len(m.senderCounts) != 0 {
		t.Fatalf(
			"indexes inconsistent after overlapping removal: hashes=%d nonces=%d senders=%d",
			len(m.hashes),
			len(m.senderNonces),
			len(m.senderCounts),
		)
	}
}

func TestMempoolRemoveExpiredMixedWorkload(t *testing.T) {
	m := NewMempool()

	const total = 25000
	const expiredCount = 12500

	now := time.Now().UTC()
	txs := make([]Transaction, total)

	for i := range txs {
		ts := now
		if i < expiredCount {
			ts = now.Add(-MempoolTransactionTTL - time.Second)
		}

		txs[i] = Transaction{
			Hash:      fmt.Sprintf("opt17k-mixed-%d", i),
			From:      fmt.Sprintf("opt17k-sender-%d", i),
			Nonce:     uint64(i + 1),
			Timestamp: ts,
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("admit tx %d: %v", i, err)
		}
	}

	m.RemoveExpiredTransactions()

	if got := m.Count(); got != total-expiredCount {
		t.Fatalf("remaining transaction count = %d, want %d",
			got, total-expiredCount)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.hashes) != total-expiredCount {
		t.Fatalf("hash index = %d, want %d",
			len(m.hashes), total-expiredCount)
	}

	if len(m.senderNonces) != total-expiredCount {
		t.Fatalf("sender nonce index = %d, want %d",
			len(m.senderNonces), total-expiredCount)
	}

	if len(m.senderCounts) != total-expiredCount {
		t.Fatalf("sender count index = %d, want %d",
			len(m.senderCounts), total-expiredCount)
	}

	for i := 0; i < expiredCount; i++ {
		if _, exists := m.hashes[txs[i].Hash]; exists {
			t.Fatalf("expired transaction still indexed: %s", txs[i].Hash)
		}
	}

	for i := expiredCount; i < total; i++ {
		if _, exists := m.hashes[txs[i].Hash]; !exists {
			t.Fatalf("live transaction missing from hash index: %s", txs[i].Hash)
		}
	}
}

func TestMempoolRemoveExpiredMixedWorkloadConcurrent(t *testing.T) {
	m := NewMempool()

	const total = 10000
	const expiredCount = 5000

	now := time.Now().UTC()
	txs := make([]Transaction, total)

	for i := range txs {
		ts := now
		if i < expiredCount {
			ts = now.Add(-MempoolTransactionTTL - time.Second)
		}

		txs[i] = Transaction{
			Hash:      fmt.Sprintf("opt17k-concurrent-%d", i),
			From:      fmt.Sprintf("opt17k-concurrent-sender-%d", i),
			Nonce:     uint64(i + 1),
			Timestamp: ts,
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("admit tx %d: %v", i, err)
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		m.RemoveExpiredTransactions()
	}()

	go func() {
		defer wg.Done()
		m.RemoveExpiredTransactions()
	}()

	wg.Wait()

	if got := m.Count(); got != total-expiredCount {
		t.Fatalf("remaining transaction count = %d, want %d",
			got, total-expiredCount)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.hashes) != total-expiredCount ||
		len(m.senderNonces) != total-expiredCount ||
		len(m.senderCounts) != total-expiredCount {
		t.Fatalf(
			"indexes inconsistent: hashes=%d nonces=%d senders=%d",
			len(m.hashes),
			len(m.senderNonces),
			len(m.senderCounts),
		)
	}
}

func TestMempoolIndexInvariantsAfterMixedRemoval(t *testing.T) {
	m := NewMempool()

	const total = 1000
	txs := make([]Transaction, total)

	for i := range txs {
		txs[i] = Transaction{
			Hash:      fmt.Sprintf("opt17m1-hash-%d", i),
			From:      fmt.Sprintf("opt17m1-sender-%d", i),
			Nonce:     uint64(i + 1),
			Timestamp: time.Now().UTC(),
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("admit tx %d: %v", i, err)
		}
	}

	// Remove half as processed.
	m.RemoveProcessedTransactions(txs[:500])

	// Make the remaining half expired.
	m.mu.Lock()
	for i := range m.Transactions {
		m.Transactions[i].Timestamp =
			time.Now().UTC().Add(-MempoolTransactionTTL - time.Second)
	}
	m.mu.Unlock()

	m.RemoveExpiredTransactions()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Transactions) != 0 {
		t.Fatalf("transactions not empty: %d", len(m.Transactions))
	}

	if len(m.hashes) != 0 {
		t.Fatalf("hash index not empty: %d", len(m.hashes))
	}

	if len(m.senderNonces) != 0 {
		t.Fatalf("sender nonce index not empty: %d", len(m.senderNonces))
	}

	if len(m.senderCounts) != 0 {
		t.Fatalf("sender count index not empty: %d", len(m.senderCounts))
	}
}

func TestMempoolIndexInvariantReAdmissionAfterMixedRemoval(t *testing.T) {
	m := NewMempool()

	const total = 1200
	txs := make([]Transaction, total)

	for i := range txs {
		txs[i] = Transaction{
			Hash:      fmt.Sprintf("opt17n2-hash-%d", i),
			From:      fmt.Sprintf("opt17n2-sender-%d", i),
			Nonce:     uint64(i + 1),
			Timestamp: time.Now().UTC(),
		}

		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("initial admission %d failed: %v", i, err)
		}
	}

	// Remove the first third as processed.
	m.RemoveProcessedTransactions(txs[:400])

	// Expire the next third.
	m.mu.Lock()
	for i := range m.Transactions {
		if i < 400 {
			m.Transactions[i].Timestamp =
				time.Now().UTC().Add(-MempoolTransactionTTL - time.Second)
		}
	}
	m.mu.Unlock()

	m.RemoveExpiredTransactions()

	m.mu.RLock()
	remaining := len(m.Transactions)
	hashes := len(m.hashes)
	nonces := len(m.senderNonces)
	senders := len(m.senderCounts)
	m.mu.RUnlock()

	if remaining != hashes || remaining != nonces {
		t.Fatalf(
			"index cardinality mismatch after mixed removal: tx=%d hashes=%d nonces=%d",
			remaining, hashes, nonces,
		)
	}

	if remaining != 400 || senders != 400 {
		t.Fatalf(
			"unexpected remaining state: tx=%d senders=%d want tx=400 senders=400",
			remaining, senders,
		)
	}

	// Re-admit the removed transactions. This verifies that processed/expired
	// removal released every admission index correctly.
	for i := 0; i < 800; i++ {
		if err := m.AdmitTransaction(txs[i]); err != nil {
			t.Fatalf("re-admission %d failed: %v", i, err)
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.Transactions) != 1200 {
		t.Fatalf("unexpected transaction count after re-admission: %d", len(m.Transactions))
	}

	if len(m.hashes) != 1200 {
		t.Fatalf("unexpected hash index count after re-admission: %d", len(m.hashes))
	}

	if len(m.senderNonces) != 1200 {
		t.Fatalf("unexpected sender nonce index count after re-admission: %d", len(m.senderNonces))
	}

	if len(m.senderCounts) != 1200 {
		t.Fatalf("unexpected sender count index count after re-admission: %d", len(m.senderCounts))
	}
}

func TestMempoolConcurrentMixedLifecycleInvariant(t *testing.T) {
	m := NewMempool()

	const initial = 1200
	initialTxs := make([]Transaction, initial)

	for i := range initialTxs {
		initialTxs[i] = Transaction{
			Hash:      fmt.Sprintf("opt17n3-initial-hash-%d", i),
			From:      fmt.Sprintf("opt17n3-initial-sender-%d", i),
			Nonce:     uint64(i + 1),
			Timestamp: time.Now().UTC(),
		}

		if err := m.AdmitTransaction(initialTxs[i]); err != nil {
			t.Fatalf("initial admission %d failed: %v", i, err)
		}
	}

	var wg sync.WaitGroup

	// Concurrent processed removal.
	wg.Add(1)
	go func() {
		defer wg.Done()

		m.RemoveProcessedTransactions(initialTxs[:300])
	}()

	// Concurrent expiry cleanup.
	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(time.Millisecond)

		m.mu.Lock()
		for i := range m.Transactions {
			if i >= 300 && i < 600 {
				m.Transactions[i].Timestamp =
					time.Now().UTC().Add(-MempoolTransactionTTL - time.Second)
			}
		}
		m.mu.Unlock()

		m.RemoveExpiredTransactions()
	}()

	// Concurrent admission of new transactions.
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < 300; i++ {
			tx := Transaction{
				Hash:      fmt.Sprintf("opt17n3-new-hash-%d", i),
				From:      fmt.Sprintf("opt17n3-new-sender-%d", i),
				Nonce:     uint64(i + 1),
				Timestamp: time.Now().UTC(),
			}

			if err := m.AdmitTransaction(tx); err != nil {
				t.Errorf("concurrent new admission %d failed: %v", i, err)
				return
			}
		}
	}()

	wg.Wait()

	m.mu.RLock()
	defer m.mu.RUnlock()

	txCount := len(m.Transactions)
	hashCount := len(m.hashes)
	nonceCount := len(m.senderNonces)
	senderCount := len(m.senderCounts)

	if txCount != hashCount || txCount != nonceCount {
		t.Fatalf(
			"mixed lifecycle index mismatch: tx=%d hashes=%d nonces=%d",
			txCount,
			hashCount,
			nonceCount,
		)
	}

	var countedSenders uint64
	for _, count := range m.senderCounts {
		countedSenders += count
	}

	if countedSenders != uint64(txCount) {
		t.Fatalf(
			"sender count invariant broken: tx=%d senderCountsTotal=%d",
			txCount,
			countedSenders,
		)
	}

	if senderCount > txCount {
		t.Fatalf(
			"sender index cardinality impossible: senders=%d tx=%d",
			senderCount,
			txCount,
		)
	}
}

func TestMempoolZeroValueAdmissionInitializesIndexes(t *testing.T) {
	var m Mempool

	tx := Transaction{
		Hash:      "opt17m1-zero-value",
		From:      "opt17m1-zero-sender",
		Nonce:     1,
		Timestamp: time.Now().UTC(),
	}

	if err := m.AdmitTransaction(tx); err != nil {
		t.Fatalf("zero-value mempool admission failed: %v", err)
	}

	if m.Count() != 1 {
		t.Fatalf("expected 1 transaction, got %d", m.Count())
	}

	m.RemoveProcessedTransactions([]Transaction{tx})

	if m.Count() != 0 {
		t.Fatalf("expected empty mempool after removal, got %d", m.Count())
	}
}
