package app

import (
	"fmt"
	"testing"
	"time"
)

func benchmarkMempoolTransaction(i int) Transaction {
	sender := i / 128

	return Transaction{
		ID:        benchmarkTransactionID(i),
		Hash:      benchmarkTransactionHash(i),
		From:      benchmarkSender(sender),
		To:        benchmarkBenchmarkTo,
		Amount:    uint64(i + 1),
		Fee:       uint64((i % 1000) + 1),
		Nonce:     uint64(i%128 + 1),
		Timestamp: time.Unix(int64(i), 0).UTC(),
	}
}

const benchmarkBenchmarkTo = "0x2222222222222222222222222222222222222222"

func benchmarkTransactionID(i int) string {
	return "bench-tx-" + benchmarkDecimal(i)
}

func benchmarkTransactionHash(i int) string {
	return benchmarkHex64(i)
}

func benchmarkSender(sender int) string {
	return "0x" + benchmarkHex40(sender+1)
}

func benchmarkDecimal(v int) string {
	if v == 0 {
		return "0"
	}

	var buf [20]byte
	pos := len(buf)

	for v > 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}

	return string(buf[pos:])
}

func benchmarkHex64(v int) string {
	const hex = "0123456789abcdef"

	var buf [64]byte
	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = hex[v&0xf]
		v >>= 4
	}

	return string(buf[:])
}

func benchmarkHex40(v int) string {
	const hex = "0123456789abcdef"

	var buf [40]byte
	for i := len(buf) - 1; i >= 0; i-- {
		buf[i] = hex[v&0xf]
		v >>= 4
	}

	return string(buf[:])
}

func benchmarkMempoolAdd(b *testing.B, size int) {
	b.Helper()

	b.StopTimer()

	txs := make([]Transaction, size)
	for i := range txs {
		txs[i] = benchmarkMempoolTransaction(i)
	}

	// Initialize once outside the timed section.
	mempool := NewMempool()

	b.StartTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, tx := range txs {
			mempool.AddTransaction(tx)
		}

		// Recreate the mempool only between benchmark iterations.
		// Keep initialization outside the timed admission path.
		b.StopTimer()
		mempool = NewMempool()
		b.StartTimer()
	}
}

func BenchmarkMempoolAdd100(b *testing.B) {
	benchmarkMempoolAdd(b, 100)
}

func BenchmarkMempoolAdd1000(b *testing.B) {
	benchmarkMempoolAdd(b, 1000)
}

func BenchmarkMempoolAdd5000(b *testing.B) {
	benchmarkMempoolAdd(b, 5000)
}

func BenchmarkMempoolAdd10000(b *testing.B) {
	benchmarkMempoolAdd(b, 10000)
}

func BenchmarkMempoolAdd25000(b *testing.B) {
	benchmarkMempoolAdd(b, 25000)
}

func benchmarkMempoolAdmit(b *testing.B, size int) {
	b.Helper()

	b.StopTimer()

	txs := make([]Transaction, size)
	for i := range txs {
		txs[i] = benchmarkMempoolTransaction(i)
	}

	mempool := NewMempool()

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StartTimer()

		for _, tx := range txs {
			if err := mempool.AdmitTransaction(tx); err != nil {
				b.Fatal(err)
			}
		}

		b.StopTimer()
		mempool = NewMempool()
	}
}

func BenchmarkMempoolAdmit25000(b *testing.B) {
	benchmarkMempoolAdmit(b, 25000)
}

func BenchmarkMempoolAdmitBatch1000(b *testing.B) {
	benchmarkMempoolAdmitBatch(b, 1000)
}

func BenchmarkMempoolAdmitBatch5000(b *testing.B) {
	benchmarkMempoolAdmitBatch(b, 5000)
}

func BenchmarkMempoolAdmitBatch25000(b *testing.B) {
	benchmarkMempoolAdmitBatch(b, 25000)
}

func BenchmarkMempoolAdmitBatch50000(b *testing.B) {
	benchmarkMempoolAdmitBatch(b, 50000)
}

func benchmarkMempoolAdmitBatch(b *testing.B, size int) {
	b.Helper()

	b.StopTimer()

	txs := make([]Transaction, size)
	for i := range txs {
		txs[i] = benchmarkMempoolTransaction(i)
	}

	b.ReportAllocs()

	mempool := NewMempool()

	for i := 0; i < b.N; i++ {
		mempool.Transactions = mempool.Transactions[:0]
		clear(mempool.hashes)
		clear(mempool.senderNonces)
		clear(mempool.senderCounts)

		b.StartTimer()

		if err := mempool.AdmitTransactions(txs); err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
	}
}

func BenchmarkMempoolAdmitBatch50000Unsorted(b *testing.B) {
	b.StopTimer()

	const size = 50000

	txs := make([]Transaction, size)
	for i := range txs {
		txs[i] = benchmarkMempoolTransaction(i)
	}

	// Reverse the batch so the admission path cannot use
	// the already-ordered fast path.
	for i, j := 0, len(txs)-1; i < j; i, j = i+1, j-1 {
		txs[i], txs[j] = txs[j], txs[i]
	}

	mempool := NewMempool()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mempool.Transactions = mempool.Transactions[:0]
		clear(mempool.hashes)
		clear(mempool.senderNonces)
		clear(mempool.senderCounts)

		b.StartTimer()

		if err := mempool.AdmitTransactions(txs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMempoolAdmitBatch25000Unsorted(b *testing.B) {
	b.StopTimer()

	const size = 25000

	txs := make([]Transaction, size)
	for i := range txs {
		txs[i] = benchmarkMempoolTransaction(i)
	}

	// Reverse the batch so the admission path cannot use
	// the already-ordered fast path.
	for i, j := 0, len(txs)-1; i < j; i, j = i+1, j-1 {
		txs[i], txs[j] = txs[j], txs[i]
	}

	mempool := NewMempool()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		mempool.Transactions = mempool.Transactions[:0]
		clear(mempool.hashes)
		clear(mempool.senderNonces)
		clear(mempool.senderCounts)

		b.StartTimer()

		if err := mempool.AdmitTransactions(txs); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMempoolRemoveProcessed10K(b *testing.B) {
	m := NewMempool()
	txs := make([]Transaction, 10000)

	for i := range txs {
		txs[i] = Transaction{
			Hash:  fmt.Sprintf("remove-10k-%d", i),
			From:  fmt.Sprintf("sender-%d", i%100),
			Nonce: uint64(i + 1),
		}
		if err := m.AdmitTransaction(txs[i]); err != nil {
			b.Fatal(err)
		}
	}

	original := append([]Transaction(nil), m.Transactions...)
	processed := txs[:5000]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Transactions = original
		m.ensureIndexesLocked()

		for k := range m.hashes {
			delete(m.hashes, k)
		}
		for k := range m.senderNonces {
			delete(m.senderNonces, k)
		}
		clear(m.senderCounts)

		for _, tx := range original {
			m.hashes[tx.Hash] = struct{}{}
			m.senderNonces[mempoolSenderNonceKey{
				From:  tx.From,
				Nonce: tx.Nonce,
			}] = struct{}{}
			m.senderCounts[tx.From]++
		}

		m.RemoveProcessedTransactions(processed)
	}
}

func BenchmarkMempoolRemoveProcessed25K(b *testing.B) {
	m := NewMempool()
	txs := make([]Transaction, 25000)

	for i := range txs {
		txs[i] = Transaction{
			Hash:  fmt.Sprintf("remove-25k-%d", i),
			From:  fmt.Sprintf("sender-%d", i%250),
			Nonce: uint64(i + 1),
		}
		if err := m.AdmitTransaction(txs[i]); err != nil {
			b.Fatal(err)
		}
	}

	original := append([]Transaction(nil), m.Transactions...)
	processed := txs[:12500]

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Transactions = original
		m.ensureIndexesLocked()

		for k := range m.hashes {
			delete(m.hashes, k)
		}
		for k := range m.senderNonces {
			delete(m.senderNonces, k)
		}
		clear(m.senderCounts)

		for _, tx := range original {
			m.hashes[tx.Hash] = struct{}{}
			m.senderNonces[mempoolSenderNonceKey{
				From:  tx.From,
				Nonce: tx.Nonce,
			}] = struct{}{}
			m.senderCounts[tx.From]++
		}

		m.RemoveProcessedTransactions(processed)
	}
}

func BenchmarkMempoolRemoveExpired25K(b *testing.B) {
	m := NewMempool()
	txs := make([]Transaction, 25000)

	for i := range txs {
		txs[i] = Transaction{
			Hash:      fmt.Sprintf("expire-25k-%d", i),
			From:      fmt.Sprintf("sender-%d", i%250),
			Nonce:     uint64(i + 1),
			Timestamp: time.Now().UTC().Add(-MempoolTransactionTTL - time.Second),
		}
		if err := m.AdmitTransaction(txs[i]); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.RemoveExpiredTransactions()
	}
}
