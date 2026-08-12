package app

import (
	"sync"
	"time"
)

const MaxMempoolTransactions = 10000
const MempoolTransactionTTL = 30 * time.Minute

type Mempool struct {
	mu           sync.RWMutex
	Transactions []Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		Transactions: make([]Transaction, 0),
	}
}

// AddTransaction adds a transaction if it is not already present and
// the mempool has not reached its capacity.
func (m *Mempool) AddTransaction(tx Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.Transactions {
		if existing.Hash == tx.Hash {
			return
		}
	}

	if len(m.Transactions) >= MaxMempoolTransactions {
		return
	}

	m.Transactions = append(m.Transactions, tx)

	m.sortByPriorityLocked()
}

// Snapshot returns an immutable copy of the current mempool contents.
// Callers may safely inspect or modify the returned slice.
func (m *Mempool) Snapshot() []Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make([]Transaction, len(m.Transactions))
	copy(snapshot, m.Transactions)

	return snapshot
}

func (m *Mempool) RemoveProcessedTransactions(processed []Transaction) {
	if len(processed) == 0 {
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

	remaining := make([]Transaction, 0, len(m.Transactions))

	for _, tx := range m.Transactions {
		if _, processed := processedMap[tx.Hash]; !processed {
			remaining = append(remaining, tx)
		}
	}

	m.Transactions = remaining
}

func (m *Mempool) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.Transactions)
}

func (m *Mempool) RemoveExpiredTransactions() {
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	remaining := make([]Transaction, 0, len(m.Transactions))

	for _, tx := range m.Transactions {
		if now.Sub(tx.Timestamp) <= MempoolTransactionTTL {
			remaining = append(remaining, tx)
		}
	}

	m.Transactions = remaining
}

func (m *Mempool) sortByPriorityLocked() {
	for i := 1; i < len(m.Transactions); i++ {
		current := m.Transactions[i]
		j := i - 1

		for j >= 0 && transactionHigherPriority(current, m.Transactions[j]) {
			m.Transactions[j+1] = m.Transactions[j]
			j--
		}

		m.Transactions[j+1] = current
	}
}

func transactionHigherPriority(a, b Transaction) bool {
	if a.Fee != b.Fee {
		return a.Fee > b.Fee
	}

	if !a.Timestamp.Equal(b.Timestamp) {
		return a.Timestamp.Before(b.Timestamp)
	}

	// Deterministic final tie-breaker.
	// Equal-fee/equal-time transactions must have
	// identical ordering on every node.
	return a.Hash < b.Hash
}
