package app

import "time"

const MaxMempoolTransactions = 10000
const MempoolTransactionTTL = 30 * time.Minute

type Mempool struct {
	Transactions []Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		Transactions: []Transaction{},
	}
}

func (m *Mempool) AddTransaction(tx Transaction) {

	for _, existing := range m.Transactions {
		if existing.Hash == tx.Hash {
			return
		}
	}
	if len(m.Transactions) >= MaxMempoolTransactions {
		return
	}

	m.Transactions = append(m.Transactions, tx)

	// Sort by priority
	m.SortByPriority()
}
func (m *Mempool) RemoveProcessedTransactions(processed []Transaction) {

	processedMap := make(map[string]bool)

	for _, tx := range processed {
		processedMap[tx.Hash] = true
	}

	var remaining []Transaction

	for _, tx := range m.Transactions {

		if !processedMap[tx.Hash] {
			remaining = append(remaining, tx)
		}
	}

	m.Transactions = remaining
}

func (m *Mempool) Count() int {
	return len(m.Transactions)
}

func (m *Mempool) RemoveExpiredTransactions() {

	now := time.Now()

	var remaining []Transaction

	for _, tx := range m.Transactions {

		if now.Sub(tx.Timestamp) <= MempoolTransactionTTL {
			remaining = append(remaining, tx)
		}
	}

	m.Transactions = remaining
}
