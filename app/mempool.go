package app

type Mempool struct {
	Transactions []Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		Transactions: []Transaction{},
	}
}

func (m *Mempool) AddTransaction(tx Transaction) {
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
