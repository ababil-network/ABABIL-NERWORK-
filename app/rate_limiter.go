package app

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]time.Time
}

var NodeRateLimiter = &RateLimiter{
	clients: make(map[string]time.Time),
}

func (r *RateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	last, ok := r.clients[ip]
	if ok && now.Sub(last) < time.Second {
		return false
	}

	r.clients[ip] = now
	return true
}
