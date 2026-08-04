package app

import "sync"

type PeerReputationManager struct {
	mu     sync.RWMutex
	scores map[string]int
}

var NodeReputation = &PeerReputationManager{
	scores: make(map[string]int),
}

func (p *PeerReputationManager) GetScore(ip string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	score, ok := p.scores[ip]
	if !ok {
		return 100
	}

	return score
}

func (p *PeerReputationManager) SetScore(ip string, score int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.scores[ip] = score
}

func (p *PeerReputationManager) Increase(ip string, value int) {
	score := p.GetScore(ip)

	score += value
	if score > 100 {
		score = 100
	}

	p.SetScore(ip, score)
}

func (p *PeerReputationManager) Decrease(ip string, value int) {
	score := p.GetScore(ip)

	score -= value
	if score < 0 {
		score = 0
	}

	p.SetScore(ip, score)
}
