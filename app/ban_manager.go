package app

import (
	"sync"
	"time"
)

type BanManager struct {
	mu         sync.Mutex
	banned     map[string]time.Time
	violations map[string]uint8
}

var NodeBan = &BanManager{
	banned:     make(map[string]time.Time),
	violations: make(map[string]uint8),
}

func (b *BanManager) Ban(ip string, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.banned[ip] = time.Now().Add(duration)
}

func (b *BanManager) IsBanned(ip string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	expire, ok := b.banned[ip]
	if !ok {
		return false
	}

	if time.Now().After(expire) {
		delete(b.banned, ip)
		return false
	}

	return true
}

func (b *BanManager) BanDuration(ip string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.violations[ip]++

	switch b.violations[ip] {
	case 1:
		return 0
	case 2:
		return 0
	case 3:
		return 0
	case 4:
		return 30 * time.Second
	case 5:
		return 2 * time.Minute
	case 6:
		return 10 * time.Minute
	case 7:
		return 30 * time.Minute
	case 8:
		return 1 * time.Hour
	case 9:
		return 12 * time.Hour
	default:
		return 24 * time.Hour
	}
}
func (b *BanManager) HandleViolation(ip string) time.Duration {
	duration := b.BanDuration(ip)

	switch duration {
	case 0:
		NodeReputation.Decrease(ip, 2)

	case 30 * time.Second:
		NodeReputation.Decrease(ip, 5)

	case 2 * time.Minute:
		NodeReputation.Decrease(ip, 10)

	case 10 * time.Minute:
		NodeReputation.Decrease(ip, 15)

	case 30 * time.Minute:
		NodeReputation.Decrease(ip, 20)

	case 1 * time.Hour:
		NodeReputation.Decrease(ip, 25)

	case 12 * time.Hour:
		NodeReputation.Decrease(ip, 35)

	default:
		NodeReputation.Decrease(ip, 50)
	}

	if duration > 0 {
		b.Ban(ip, duration)
	}

	return duration
}
