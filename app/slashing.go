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

	for i := range Validators {

		if Validators[i].Address == address {

			Validators[i].Jailed = true
			Validators[i].Active = false

			return true
		}
	}

	return false
}

func UnjailValidator(address string) bool {

	for i := range Validators {

		if Validators[i].Address == address {

			Validators[i].Jailed = false
			Validators[i].Active = true
			Validators[i].MissedBlocks = 0

			return true
		}
	}

	return false
}
