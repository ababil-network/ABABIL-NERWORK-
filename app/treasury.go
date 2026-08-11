package app

import (
	"errors"
	"math"
	"time"
)

var (
	errTreasuryOverflow       = errors.New("treasury balance overflow")
	errTreasuryRecordOverflow = errors.New("treasury record amount overflow")
)

type Treasury struct {
	Ecosystem uint64
	Security  uint64
}

type TreasuryRecord struct {
	ID     uint64
	Type   string
	Amount uint64
	Reason string
	Block  uint64
	Time   time.Time
}

var NetworkTreasury Treasury
var TreasuryHistory []TreasuryRecord

// DepositTreasury adds ecosystem and security funds atomically.
// If any balance or record amount would overflow, nothing is changed.
func DepositTreasury(ecosystem uint64, security uint64) error {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	if ecosystem > math.MaxUint64-NetworkTreasury.Ecosystem {
		return errTreasuryOverflow
	}

	if security > math.MaxUint64-NetworkTreasury.Security {
		return errTreasuryOverflow
	}

	if ecosystem > math.MaxUint64-security {
		return errTreasuryRecordOverflow
	}

	if uint64(len(TreasuryHistory)) == math.MaxUint64 {
		return errTreasuryRecordOverflow
	}

	NetworkTreasury.Ecosystem += ecosystem
	NetworkTreasury.Security += security

	TreasuryHistory = append(TreasuryHistory, TreasuryRecord{
		ID:     uint64(len(TreasuryHistory)) + 1,
		Type:   "Deposit",
		Amount: ecosystem + security,
		Reason: "Reward Distribution",
		Block:  0,
		Time:   time.Now(),
	})

	return nil
}

func GetTreasuryBalance() Treasury {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	return NetworkTreasury
}
