package app

import (
	"sync"
	"time"
)

const DailyFreeTransactionLimit uint64 = 1000

type FreeTransactionInfo struct {
	Count uint64
	Day   string
}

type FreeTransactionManager struct {
	mu   sync.RWMutex
	data map[string]*FreeTransactionInfo
}

var NodeFreeTransaction = &FreeTransactionManager{
	data: make(map[string]*FreeTransactionInfo),
}

func currentDay() string {
	return time.Now().UTC().Format("2006-01-02")
}

func (m *FreeTransactionManager) Remaining(address string) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	today := currentDay()

	info, ok := m.data[address]
	if !ok {
		m.data[address] = &FreeTransactionInfo{
			Count: 0,
			Day:   today,
		}
		return WalletFreeLimit(address)
	}

	if info.Day != today {
		info.Day = today
		info.Count = 0
	}

	if info.Count >= DailyFreeTransactionLimit {
		return 0
	}

        limit := WalletFreeLimit(address)

if info.Count >= limit {
	return 0
}

return limit - info.Count
}

func (m *FreeTransactionManager) Use(address string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	today := currentDay()

	info, ok := m.data[address]
	if !ok {
		info = &FreeTransactionInfo{
			Count: 0,
			Day:   today,
		}
		m.data[address] = info
	}

	if info.Day != today {
		info.Day = today
		info.Count = 0
	}

	limit := WalletFreeLimit(address)

if info.Count >= limit {
	return false
	}

	info.Count++

	return true
}
