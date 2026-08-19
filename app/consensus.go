package app

import "sync"

type Validator struct {
	// ID is permanent and must never change after registration.
	ID uint64

	// Slot is the validator's current active position.
	// Slot is dynamic and may change after validator-set compression.
	Slot uint64

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

// nextValidatorIDLocked returns the next permanent validator ID.
// IDs are never reused, even when validators exit.
func nextValidatorIDLocked() uint64 {
	var maxID uint64

	for _, v := range Validators {
		if v.ID > maxID {
			maxID = v.ID
		}
	}

	return maxID + 1
}

// nextValidatorSlotLocked returns the next available active slot.
// Slots are dynamic and are assigned from the current validator count.
func nextValidatorSlotLocked() uint64 {
	var maxSlot uint64

	for _, v := range Validators {
		if v.Slot > maxSlot {
			maxSlot = v.Slot
		}
	}

	return maxSlot + 1
}

// normalizeValidatorSlotsLocked compresses active validator slots.
// Permanent validator IDs are never modified.
//
// Validators that remain active receive contiguous slots starting from 1.
// Inactive/jailed validators keep their permanent ID but do not occupy an
// active slot.
//
// This function must only be called while validatorStateMu is held.
func normalizeValidatorSlotsLocked() {
	activeSlot := uint64(1)

	for i := range Validators {
		if Validators[i].Active && !Validators[i].Jailed {
			Validators[i].Slot = activeSlot
			activeSlot++
		}
	}
}

// ValidatorSlot returns the current dynamic slot for a validator ID.
func ValidatorSlot(id uint64) (uint64, bool) {
	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	for _, v := range Validators {
		if v.ID == id {
			return v.Slot, v.Active && !v.Jailed
		}
	}

	return 0, false
}

// CompressValidatorSlots atomically rebuilds the active validator slots.
//
// Permanent validator IDs remain unchanged. Only active, non-jailed
// validators receive contiguous dynamic slots.
func compressValidatorSlotsLocked() error {
	originalValidators := append([]Validator(nil), Validators...)
	originalLeaderIndex := LeaderIndex

	normalizeValidatorSlotsLocked()

	if len(Validators) == 0 {
		LeaderIndex = 0
		return nil
	}

	if LeaderIndex < 0 || LeaderIndex >= len(Validators) ||
		!Validators[LeaderIndex].Active || Validators[LeaderIndex].Jailed {
		findEligibleLeaderLocked(LeaderIndex)
	}

	if err := reconcileValidatorCollateralsLocked(); err != nil {
		Validators = originalValidators
		LeaderIndex = originalLeaderIndex
		return err
	}

	return nil
}

// CompressValidatorSlots atomically rebuilds the active validator slots.
//
// Permanent validator IDs remain unchanged. Only active, non-jailed
// validators receive contiguous dynamic slots.
//
// Slot compression and collateral reconciliation succeed or fail together.
func CompressValidatorSlots() {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	if err := compressValidatorSlotsLocked(); err != nil {
		LogInfo("Validator slot compression rejected: " + err.Error())
	}
}

func addValidatorLocked(address string, power uint64) {
	for _, v := range Validators {
		if v.Address == address {
			return
		}
	}

	Validators = append(Validators, Validator{
		ID:           nextValidatorIDLocked(),
		Slot:         nextValidatorSlotLocked(),
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
