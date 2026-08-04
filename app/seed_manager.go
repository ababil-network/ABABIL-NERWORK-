package app

import "sync"

type SeedManager struct {
	mu    sync.RWMutex
	seeds []string
}

var NodeSeed = &SeedManager{}

func (s *SeedManager) Add(address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, seed := range s.seeds {
		if seed == address {
			return
		}
	}

	s.seeds = append(s.seeds, address)
}

func (s *SeedManager) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, len(s.seeds))
	copy(result, s.seeds)

	return result
}

func (s *SeedManager) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.seeds)
}
