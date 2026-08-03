package app

import "sync"

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

	n.nonces[address]++

	return n.nonces[address]
}

func (n *NonceManager) Verify(address string, nonce uint64) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	current := n.nonces[address]

	return nonce == current+1
}
func (n *NonceManager) Reset(address string) {

	n.mu.Lock()
	defer n.mu.Unlock()

	n.nonces[address] = 0
}
