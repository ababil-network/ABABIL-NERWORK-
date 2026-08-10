package app

// SortByPriority orders the mempool from highest fee to lowest fee.
// Transactions with equal fees are ordered oldest-first.
func (m *Mempool) SortByPriority() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sortByPriorityLocked()
}
