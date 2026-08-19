package app

import "time"

type SlashRecord struct {
	ID        uint64
	Validator string
	Reason    string
	Penalty   uint64
	Block     uint64
	Time      time.Time
}

var SlashHistory []SlashRecord

func JailValidator(address string) bool {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	for i := range Validators {
		if Validators[i].Address != address {
			continue
		}

		originalValidators := append([]Validator(nil), Validators...)
		originalLeaderIndex := LeaderIndex

		// A jailed validator immediately leaves the active validator set.
		// Its permanent ID remains unchanged.
		Validators[i].Jailed = true
		Validators[i].Active = false
		Validators[i].Slot = 0

		if err := compressValidatorSlotsLocked(); err != nil {
			Validators = originalValidators
			LeaderIndex = originalLeaderIndex
			return false
		}

		return true
	}

	return false
}

func UnjailValidator(address string) bool {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	for i := range Validators {
		if Validators[i].Address != address {
			continue
		}

		originalValidators := append([]Validator(nil), Validators...)
		originalLeaderIndex := LeaderIndex

		// Re-enter the active validator set through the normal dynamic
		// slot compression rule. Permanent ID is unchanged.
		Validators[i].Jailed = false
		Validators[i].Active = true
		Validators[i].MissedBlocks = 0

		if err := compressValidatorSlotsLocked(); err != nil {
			Validators = originalValidators
			LeaderIndex = originalLeaderIndex
			return false
		}

		return true
	}

	return false
}
