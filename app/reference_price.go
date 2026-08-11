package app

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

const (
	ReferencePriceObservationWindow = 5 * time.Minute
	ReferencePriceMaxAge            = 2 * time.Minute

	// Maximum allowed deviation from the current reference price.
	// Expressed in basis points: 1000 = 10%.
	ReferencePriceMaxDeviationBPS uint64 = 1000

	BasisPoints uint64 = 10000
)

var (
	ErrReferencePriceUnavailable = errors.New("ABABIL reference price unavailable")
	ErrReferencePriceStale       = errors.New("ABABIL reference price is stale")
	ErrReferencePriceInvalid     = errors.New("invalid ABABIL reference price")
	ErrReferencePriceDeviation   = errors.New("ABABIL reference price deviation too large")
)

type PriceObservation struct {
	PriceMicroUSD uint64
	Timestamp     time.Time
}

type ReferencePriceManager struct {
	mu sync.RWMutex

	observations []PriceObservation
	reference    uint64
	updatedAt    time.Time
}

var NodeReferencePrice = &ReferencePriceManager{}

func (r *ReferencePriceManager) AddObservation(
	priceMicroUSD uint64,
	timestamp time.Time,
) error {
	if priceMicroUSD == 0 {
		return ErrReferencePriceInvalid
	}

	if timestamp.IsZero() {
		return ErrReferencePriceInvalid
	}

	now := time.Now().UTC()

	if timestamp.After(now.Add(time.Minute)) {
		return ErrReferencePriceInvalid
	}

	if now.Sub(timestamp) > ReferencePriceObservationWindow {
		return ErrReferencePriceStale
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.observations = append(r.observations, PriceObservation{
		PriceMicroUSD: priceMicroUSD,
		Timestamp:     timestamp.UTC(),
	})

	r.removeExpiredLocked(now)

	return r.recalculateLocked()
}

func (r *ReferencePriceManager) removeExpiredLocked(now time.Time) {
	cutoff := now.Add(-ReferencePriceObservationWindow)

	remaining := make(
		[]PriceObservation,
		0,
		len(r.observations),
	)

	for _, observation := range r.observations {
		if !observation.Timestamp.Before(cutoff) {
			remaining = append(remaining, observation)
		}
	}

	r.observations = remaining
}

func (r *ReferencePriceManager) recalculateLocked() error {
	if len(r.observations) == 0 {
		return ErrReferencePriceUnavailable
	}

	values := make([]uint64, 0, len(r.observations))

	for _, observation := range r.observations {
		if observation.PriceMicroUSD == 0 {
			continue
		}

		values = append(values, observation.PriceMicroUSD)
	}

	if len(values) == 0 {
		return ErrReferencePriceUnavailable
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})

	median := values[len(values)/2]

	if r.reference != 0 {
		if !withinDeviation(
			r.reference,
			median,
			ReferencePriceMaxDeviationBPS,
		) {
			return ErrReferencePriceDeviation
		}
	}

	r.reference = median
	r.updatedAt = time.Now().UTC()

	return nil
}

func withinDeviation(
	oldPrice uint64,
	newPrice uint64,
	maxDeviationBPS uint64,
) bool {
	if oldPrice == 0 {
		return false
	}

	var difference uint64

	if newPrice >= oldPrice {
		difference = newPrice - oldPrice
	} else {
		difference = oldPrice - newPrice
	}

	if difference > math.MaxUint64/BasisPoints {
		return false
	}

	return difference*BasisPoints <= oldPrice*maxDeviationBPS
}

func (r *ReferencePriceManager) Price() (uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.reference == 0 || r.updatedAt.IsZero() {
		return 0, ErrReferencePriceUnavailable
	}

	if time.Since(r.updatedAt) > ReferencePriceMaxAge {
		return 0, ErrReferencePriceStale
	}

	return r.reference, nil
}

func (r *ReferencePriceManager) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.observations = nil
	r.reference = 0
	r.updatedAt = time.Time{}
}
