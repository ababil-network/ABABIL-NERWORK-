package app

import (
	"sync"
	"time"
)

type PeerHealthManager struct {
	mu       sync.RWMutex
	lastSeen map[string]time.Time
}

var NodeHealth = &PeerHealthManager{
	lastSeen: make(map[string]time.Time),
}

func (h *PeerHealthManager) Update(ip string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.lastSeen[ip] = time.Now()
}

func (h *PeerHealthManager) LastSeen(ip string) time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.lastSeen[ip]
}
