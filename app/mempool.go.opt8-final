package app

import (
	"sort"
	"sync"
	"time"
)

const MaxMempoolTransactions = 50000

const MempoolTransactionTTL = 30 * time.Minute

type mempoolSenderNonceKey struct {
	From  string
	Nonce uint64
}

type Mempool struct {
	mu sync.RWMutex

	// Transactions are kept in admission order.
	Transactions []Transaction

	// O(1) duplicate hash detection.
	hashes map[string]struct{}

	// O(1) sender+nonce conflict detection.
	senderNonces map[mempoolSenderNonceKey]struct{}

	// O(1) pending transaction count per sender.
	senderCounts map[string]uint64
}

func NewMempool() *Mempool {
	return &Mempool{
		Transactions: make([]Transaction, 0, MaxMempoolTransactions),
		hashes:       make(map[string]struct{}, MaxMempoolTransactions),
		senderNonces: make(map[mempoolSenderNonceKey]struct{}, MaxMempoolTransactions),
		senderCounts: make(map[string]uint64),
	}
}

func (m *Mempool) ensureIndexesLocked() {
	if m.hashes == nil {
		m.hashes = make(map[string]struct{}, MaxMempoolTransactions)
	}

	if m.senderNonces == nil {
		m.senderNonces = make(
			map[mempoolSenderNonceKey]struct{},
			MaxMempoolTransactions,
		)
	}

	if m.senderCounts == nil {
		m.senderCounts = make(map[string]uint64)
	}

	// Rebuild indexes only when a zero-value/legacy mempool
	// actually needs reconstruction.
	if len(m.Transactions) == 0 {
		return
	}

	if len(m.hashes) == 0 &&
		len(m.senderNonces) == 0 &&
		len(m.senderCounts) == 0 {

		for _, tx := range m.Transactions {
			m.hashes[tx.Hash] = struct{}{}

			m.senderNonces[mempoolSenderNonceKey{
				From:  tx.From,
				Nonce: tx.Nonce,
			}] = struct{}{}

			m.senderCounts[tx.From]++
		}
	}
} // AddTransaction preserves the historical API.
// Invalid/duplicate/full transactions are silently ignored.
// Consensus validation must happen before admission.
func (m *Mempool) AddTransaction(tx Transaction) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureIndexesLocked()

	if len(m.Transactions) >= MaxMempoolTransactions {
		return
	}

	if _, exists := m.hashes[tx.Hash]; exists {
		return
	}

	senderNonce := mempoolSenderNonceKey{
		From:  tx.From,
		Nonce: tx.Nonce,
	}

	// Preserve the historical AddTransaction API:
	// sender+nonce conflicts are enforced by AdmitTransaction(),
	// not by this legacy insertion method.
	m.Transactions = append(m.Transactions, tx)

	m.hashes[tx.Hash] = struct{}{}
	m.senderNonces[senderNonce] = struct{}{}
	m.senderCounts[tx.From]++
}

// Snapshot returns a defensive copy ordered by priority.
//
// Priority:
//  1. Higher fee first.
//  2. Older timestamp first.
//  3. Hash ascending as deterministic tie-breaker.
//
// The internal mempool order is never modified.
func (m *Mempool) Snapshot() []Transaction {
	if m == nil {
		return nil
	}

	m.mu.RLock()

	snapshot := make([]Transaction, len(m.Transactions))
	copy(snapshot, m.Transactions)

	m.mu.RUnlock()

	sort.Slice(snapshot, func(i, j int) bool {
		return transactionHigherPriority(snapshot[i], snapshot[j])
	})

	return snapshot
}

func (m *Mempool) RemoveProcessedTransactions(processed []Transaction) {
	if m == nil || len(processed) == 0 {
		return
	}

	processedMap := make(map[string]struct{}, len(processed))

	for _, tx := range processed {
		if tx.Hash != "" {
			processedMap[tx.Hash] = struct{}{}
		}
	}

	if len(processedMap) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureIndexesLocked()

	remaining := make(
		[]Transaction,
		0,
		len(m.Transactions),
	)

	for _, tx := range m.Transactions {
		if _, processed := processedMap[tx.Hash]; processed {
			delete(m.hashes, tx.Hash)

			delete(
				m.senderNonces,
				mempoolSenderNonceKey{
					From:  tx.From,
					Nonce: tx.Nonce,
				},
			)

			if count := m.senderCounts[tx.From]; count > 1 {
				m.senderCounts[tx.From] = count - 1
			} else {
				delete(m.senderCounts, tx.From)
			}

			continue
		}

		remaining = append(remaining, tx)
	}

	m.Transactions = remaining
}

func (m *Mempool) Count() int {
	if m == nil {
		return 0
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Transactions)
}

func (m *Mempool) RemoveExpiredTransactions() {
	if m == nil {
		return
	}

	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.ensureIndexesLocked()

	remaining := make(
		[]Transaction,
		0,
		len(m.Transactions),
	)

	for _, tx := range m.Transactions {
		if now.Sub(tx.Timestamp) <= MempoolTransactionTTL {
			remaining = append(remaining, tx)
			continue
		}

		delete(m.hashes, tx.Hash)

		delete(
			m.senderNonces,
			mempoolSenderNonceKey{
				From:  tx.From,
				Nonce: tx.Nonce,
			},
		)

		if count := m.senderCounts[tx.From]; count > 1 {
			m.senderCounts[tx.From] = count - 1
		} else {
			delete(m.senderCounts, tx.From)
		}
	}

	m.Transactions = remaining
}

func (m *Mempool) sortByPriorityLocked() {
	sort.Slice(m.Transactions, func(i, j int) bool {
		return transactionHigherPriority(
			m.Transactions[i],
			m.Transactions[j],
		)
	})
}

func transactionHigherPriority(a, b Transaction) bool {
	if a.Fee != b.Fee {
		return a.Fee > b.Fee
	}

	if !a.Timestamp.Equal(b.Timestamp) {
		return a.Timestamp.Before(b.Timestamp)
	}

	return a.Hash < b.Hash
}
