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

func GetLeader() *Validator {
	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	for i := range Validators {
		if Validators[i].Active && !Validators[i].Jailed {
			return &Validators[i]
		}
	}

	return nil
}

var LeaderIndex int

func RotateLeader() *Validator {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	if len(Validators) == 0 {
		return nil
	}

	for {
		LeaderIndex++

		if LeaderIndex >= len(Validators) {
			LeaderIndex = 0
		}

		if Validators[LeaderIndex].Active &&
			!Validators[LeaderIndex].Jailed {
			return &Validators[LeaderIndex]
		}
	}
}
