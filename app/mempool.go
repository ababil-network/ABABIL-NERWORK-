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
}
