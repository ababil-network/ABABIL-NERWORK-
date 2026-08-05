package app

import "sync"

type PeerConnectionManager struct {
	mu          sync.Mutex
	connections map[string]int
}

var NodeConnections = &PeerConnectionManager{
	connections: make(map[string]int),
}

func (p *PeerConnectionManager) Allow(ip string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.connections[ip] < NodeNetworkConfig.MaxConnectionsPerIP
}

func (p *PeerConnectionManager) Add(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections[ip]++
}

func (p *PeerConnectionManager) Remove(ip string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connections[ip] > 0 {
		p.connections[ip]--
	}

	if p.connections[ip] == 0 {
		delete(p.connections, ip)
	}
}
