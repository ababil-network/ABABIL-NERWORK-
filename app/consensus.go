package app

import "sync"

type Validator struct {
	ID           uint64
	Address      string
	ConsensusKey string
	Power        uint64
	Commission   uint8
	Active       bool
	Jailed       bool
	MissedBlocks uint64
	Genesis      bool
}

var Validators []Validator

// validatorStateMu protects Validators and LeaderIndex from concurrent access.
var validatorStateMu sync.RWMutex

func AddValidator(address string, power uint64) {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	addValidatorLocked(address, power)
}

func addValidatorLocked(address string, power uint64) {
	for _, v := range Validators {
		if v.Address == address {
			return
		}
	}

	Validators = append(Validators, Validator{
		ID:           uint64(len(Validators) + 1),
		Address:      address,
		ConsensusKey: "",
		Power:        power,
		Commission:   5,
		Active:       true,
		Jailed:       false,
		MissedBlocks: 0,
		Genesis:      false,
	})
}

func AddGenesisValidator(address string, power uint64) {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	addValidatorLocked(address, power)

	if len(Validators) > 0 {
		Validators[0].Genesis = true
	}
}

// LeaderIndex identifies the current leader position in Validators.
// The effective leader is always an active, non-jailed validator.
var LeaderIndex int

// findEligibleLeaderLocked returns the first eligible validator starting at
// the supplied index. It also normalizes LeaderIndex to the selected validator.
func findEligibleLeaderLocked(start int) *Validator {
	if len(Validators) == 0 {
		return nil
	}

	if start < 0 || start >= len(Validators) {
		start = 0
	}

	for step := 0; step < len(Validators); step++ {
		candidate := (start + step) % len(Validators)

		if Validators[candidate].Active && !Validators[candidate].Jailed {
			LeaderIndex = candidate
			v := Validators[candidate]
			return &v
		}
	}

	return nil
}

// GetLeader returns the current effective leader.
//
// If LeaderIndex points to an inactive or jailed validator, the index is
// normalized to the next eligible validator. This guarantees that active
// validators are not stranded behind a jailed leader.
func GetLeader() *Validator {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	return findEligibleLeaderLocked(LeaderIndex)
}

// RotateLeader advances from the current effective leader to the next
// active, non-jailed validator.
//
// GetLeader and RotateLeader use the same eligibility rule, so they cannot
// disagree about which validators are eligible to lead.
func RotateLeader() *Validator {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	if len(Validators) == 0 {
		return nil
	}

	current := LeaderIndex
	if current < 0 || current >= len(Validators) {
		current = 0
	}

	// Find the current effective leader first.
	currentLeader := findEligibleLeaderLocked(current)
	if currentLeader == nil {
		return nil
	}

	// Advance strictly after the current leader.
	for step := 1; step <= len(Validators); step++ {
		candidate := (LeaderIndex + step) % len(Validators)

		if Validators[candidate].Active && !Validators[candidate].Jailed {
			LeaderIndex = candidate
			v := Validators[candidate]
			return &v
		}
	}

	// Only one eligible validator exists.
	return currentLeader
}
