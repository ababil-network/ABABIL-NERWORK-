package app

import (
	"errors"
	"sort"
)

var (
	ErrMempoolAdmissionDuplicateHash  = errors.New("transaction already exists in mempool")
	ErrMempoolAdmissionDuplicateNonce = errors.New("sender nonce already exists in mempool")
	ErrMempoolAdmissionSenderLimit    = errors.New("sender mempool limit reached")
	ErrMempoolAdmissionFull           = errors.New("mempool capacity reached")
)

const (
	NormalSenderPendingLimit uint64 = 256

	// maxBatchIndex16Entries is the largest batch size whose indexes
	// can safely be represented by uint16.
	//
	// A batch of exactly 65,535 entries has indexes 0..65,534.
	// The next entry count, 65,536, requires uint32.
	maxBatchIndex16Entries = int(^uint16(0))

	Congestion95SenderPendingLimit  uint64 = 128
	Congestion96SenderPendingLimit  uint64 = 102
	Congestion97SenderPendingLimit  uint64 = 76
	Congestion98SenderPendingLimit  uint64 = 51
	Congestion99SenderPendingLimit  uint64 = 25
	Congestion100SenderPendingLimit uint64 = 16
)

func useBatchIndex16(batchSize int) bool {
	return batchSize <= maxBatchIndex16Entries
}

// SenderPendingLimit returns the node-local pending transaction limit
// for one sender at the current network load.
func SenderPendingLimit(load uint64) uint64 {
	switch {
	case load < FeeLoadCongestionEndPercent:
		return NormalSenderPendingLimit

	case load == 95:
		return Congestion95SenderPendingLimit

	case load == 96:
		return Congestion96SenderPendingLimit

	case load == 97:
		return Congestion97SenderPendingLimit

	case load == 98:
		return Congestion98SenderPendingLimit

	case load == 99:
		return Congestion99SenderPendingLimit

	default:
		return Congestion100SenderPendingLimit
	}
}

// AdmitTransaction performs node-local mempool admission checks.
//
// Consensus validation must happen before this function.
//
// All admission indexes are O(1):
//   - hash lookup
//   - sender+nonce lookup
//   - sender pending count lookup
func (m *Mempool) AdmitTransaction(tx Transaction) error {
	if m == nil {
		return ErrMempoolAdmissionFull
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Transactions) >= MaxMempoolTransactions {
		return ErrMempoolAdmissionFull
	}

	if _, exists := m.hashes[tx.Hash]; exists {
		return ErrMempoolAdmissionDuplicateHash
	}

	senderNonce := mempoolSenderNonceKey{
		From:  tx.From,
		Nonce: tx.Nonce,
	}

	if _, exists := m.senderNonces[senderNonce]; exists {
		return ErrMempoolAdmissionDuplicateNonce
	}

	load := NodeDynamicFee.Load()
	senderLimit := SenderPendingLimit(load)

	senderCount := m.senderCounts[tx.From]
	if senderCount >= senderLimit {
		return ErrMempoolAdmissionSenderLimit
	}

	m.Transactions = append(m.Transactions, tx)
	m.hashes[tx.Hash] = struct{}{}
	m.senderNonces[senderNonce] = struct{}{}
	m.senderCounts[tx.From] = senderCount + 1

	return nil
}

func (m *Mempool) AdmitTransactions(txs []Transaction) error {
	if m == nil {
		return ErrMempoolAdmissionFull
	}

	if len(txs) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureIndexesLocked()

	if len(txs) > MaxMempoolTransactions-len(m.Transactions) {
		return ErrMempoolAdmissionFull
	}

	load := NodeDynamicFee.Load()
	senderLimit := SenderPendingLimit(load)

	// Validate against the existing mempool first.
	//
	// Fast path: already ordered batches do not allocate a temporary
	// index slice. Unsorted batches use compact uint16 indexes.
	sorted := true
	for i, tx := range txs {
		if _, exists := m.hashes[tx.Hash]; exists {
			return ErrMempoolAdmissionDuplicateHash
		}

		senderNonce := mempoolSenderNonceKey{
			From:  tx.From,
			Nonce: tx.Nonce,
		}

		if _, exists := m.senderNonces[senderNonce]; exists {
			return ErrMempoolAdmissionDuplicateNonce
		}

		if i > 0 {
			prev := txs[i-1]
			if prev.From > tx.From ||
				(prev.From == tx.From && prev.Nonce > tx.Nonce) {
				sorted = false
			}
		}
	}

	// Unsorted batches use the smallest safe scratch index width.
	//
	// <= 65,535 entries: uint16.
	// > 65,535 entries: uint32.
	//
	// The uint32 buffer is allocated lazily only when a batch actually
	// requires it. This keeps the normal path memory-efficient while
	// remaining safe for future 100k+ transaction batches.
	if sorted {
		for i := 0; i < len(txs); {
			tx := txs[i]
			j := i + 1

			for j < len(txs) {
				next := txs[j]

				if next.From != tx.From {
					break
				}

				if next.Nonce == tx.Nonce {
					return ErrMempoolAdmissionDuplicateNonce
				}

				j++
			}

			batchCount := uint64(j - i)

			if m.senderCounts[tx.From]+batchCount > senderLimit {
				return ErrMempoolAdmissionSenderLimit
			}

			i = j
		}
	} else if useBatchIndex16(len(txs)) {
		// Small unsorted batch: uint16 fast path.
		if cap(m.batchIndexes16) < len(txs) {
			m.batchIndexes16 = make([]uint16, len(txs))
		} else {
			m.batchIndexes16 = m.batchIndexes16[:len(txs)]
		}

		for i := range m.batchIndexes16 {
			m.batchIndexes16[i] = uint16(i)
		}

		sort.Slice(m.batchIndexes16, func(i, j int) bool {
			a := txs[m.batchIndexes16[i]]
			b := txs[m.batchIndexes16[j]]

			if a.From != b.From {
				return a.From < b.From
			}

			return a.Nonce < b.Nonce
		})

		for i := 0; i < len(m.batchIndexes16); {
			tx := txs[m.batchIndexes16[i]]
			j := i + 1

			for j < len(m.batchIndexes16) {
				next := txs[m.batchIndexes16[j]]

				if next.From != tx.From {
					break
				}

				if next.Nonce == tx.Nonce {
					return ErrMempoolAdmissionDuplicateNonce
				}

				j++
			}

			batchCount := uint64(j - i)

			if m.senderCounts[tx.From]+batchCount > senderLimit {
				return ErrMempoolAdmissionSenderLimit
			}

			i = j
		}
	} else {
		// Large unsorted batch: lazy uint32 path.
		//
		// This path is completely unused for batches <= 65,535.
		if cap(m.batchIndexes32) < len(txs) {
			m.batchIndexes32 = make([]uint32, len(txs))
		} else {
			m.batchIndexes32 = m.batchIndexes32[:len(txs)]
		}

		for i := range m.batchIndexes32 {
			m.batchIndexes32[i] = uint32(i)
		}

		sort.Slice(m.batchIndexes32, func(i, j int) bool {
			a := txs[m.batchIndexes32[i]]
			b := txs[m.batchIndexes32[j]]

			if a.From != b.From {
				return a.From < b.From
			}

			return a.Nonce < b.Nonce
		})

		for i := 0; i < len(m.batchIndexes32); {
			tx := txs[m.batchIndexes32[i]]
			j := i + 1

			for j < len(m.batchIndexes32) {
				next := txs[m.batchIndexes32[j]]

				if next.From != tx.From {
					break
				}

				if next.Nonce == tx.Nonce {
					return ErrMempoolAdmissionDuplicateNonce
				}

				j++
			}

			batchCount := uint64(j - i)

			if m.senderCounts[tx.From]+batchCount > senderLimit {
				return ErrMempoolAdmissionSenderLimit
			}

			i = j
		}
	}

	// Commit only after the entire batch passes validation.
	//
	// Hash and sender+nonce indexes remain per-transaction because they
	// provide the O(1) duplicate checks required by admission.
	for _, tx := range txs {
		senderNonce := mempoolSenderNonceKey{
			From:  tx.From,
			Nonce: tx.Nonce,
		}

		m.Transactions = append(m.Transactions, tx)
		m.hashes[tx.Hash] = struct{}{}
		m.senderNonces[senderNonce] = struct{}{}
	}

	// Update sender counts once per sender group rather than once per
	// transaction. This preserves the exact final counts while reducing
	// senderCounts map writes substantially for large batches.
	if sorted {
		for i := 0; i < len(txs); {
			tx := txs[i]
			j := i + 1

			for j < len(txs) && txs[j].From == tx.From {
				j++
			}

			m.senderCounts[tx.From] += uint64(j - i)
			i = j
		}
	} else if len(txs) <= maxBatchIndex16Entries {
		// Commit counts using the same uint16 ordering used during validation.
		for i := 0; i < len(m.batchIndexes16); {
			tx := txs[m.batchIndexes16[i]]
			j := i + 1

			for j < len(m.batchIndexes16) &&
				txs[m.batchIndexes16[j]].From == tx.From {
				j++
			}

			m.senderCounts[tx.From] += uint64(j - i)
			i = j
		}
	} else {
		// Commit counts using the uint32 ordering used for large batches.
		for i := 0; i < len(m.batchIndexes32); {
			tx := txs[m.batchIndexes32[i]]
			j := i + 1

			for j < len(m.batchIndexes32) &&
				txs[m.batchIndexes32[j]].From == tx.From {
				j++
			}

			m.senderCounts[tx.From] += uint64(j - i)
			i = j
		}
	}

	return nil
}
