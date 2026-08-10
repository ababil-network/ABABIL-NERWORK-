package app

import (
	"math"
	"sync"
)

type NonceManager struct {
	mu     sync.RWMutex
	nonces map[string]uint64
}

var NodeNonce = &NonceManager{
	nonces: make(map[string]uint64),
}

func (n *NonceManager) Get(address string) uint64 {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.nonces[address]
}

func (n *NonceManager) Set(address string, nonce uint64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nonces[address] = nonce
}

func (n *NonceManager) Next(address string) uint64 {
	n.mu.Lock()
	defer n.mu.Unlock()

	current := n.nonces[address]

	if current == math.MaxUint64 {
		return current
	}

	current++
	n.nonces[address] = current

	return current
}

func (n *NonceManager) Verify(address string, nonce uint64) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	current := n.nonces[address]

	if current == math.MaxUint64 {
		return false
	}

	return nonce == current+1
}

// TrySet atomically verifies that nonce is exactly the next nonce
// and commits it if valid.
//
// This operation must be used during transaction execution instead
// of performing Verify followed later by Set.
func (n *NonceManager) TrySet(address string, nonce uint64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	current := n.nonces[address]

	if current == math.MaxUint64 {
		return false
	}

	if nonce != current+1 {
		return false
	}

	n.nonces[address] = nonce
	return true
}

// Rollback restores the previous nonce only when the current nonce
// is exactly the reserved nonce.
//
// This prevents an older failed transaction from rolling back a
// newer transaction's nonce.
func (n *NonceManager) Rollback(address string, nonce uint64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	current := n.nonces[address]

	if current != nonce {
		return false
	}

	if nonce == 0 {
		delete(n.nonces, address)
		return true
	}

	n.nonces[address] = nonce - 1
	return true
}

func (n *NonceManager) Reset(address string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.nonces[address] = 0
}
