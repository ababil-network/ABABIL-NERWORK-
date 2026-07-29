package app

import "sync"

type ReplayManager struct {
	mu   sync.RWMutex
	seen map[string]bool
}

var NodeReplay = &ReplayManager{
	seen: make(map[string]bool),
}

// Check returns true if the transaction hash has NOT been seen before.
func (r *ReplayManager) Check(hash string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return !r.seen[hash]
}

// Add marks a transaction hash as already processed.
func (r *ReplayManager) Add(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seen[hash] = true
}

// Remove deletes a transaction hash (useful if a transaction is dropped).
func (r *ReplayManager) Remove(hash string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.seen, hash)
}

// Exists returns true if the transaction hash already exists.
func (r *ReplayManager) Exists(hash string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.seen[hash]
}
