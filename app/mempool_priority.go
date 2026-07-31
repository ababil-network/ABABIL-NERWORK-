package app

import "sort"

func (m *Mempool) SortByPriority() {
	sort.Slice(m.Transactions, func(i, j int) bool {

		// Higher fee first
		if m.Transactions[i].Fee != m.Transactions[j].Fee {
			return m.Transactions[i].Fee > m.Transactions[j].Fee
		}

		// Same fee -> older transaction first
		return m.Transactions[i].Timestamp.Before(
			m.Transactions[j].Timestamp,
		)
	})
}
