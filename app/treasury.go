package app

import "time"

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

func DepositTreasury(ecosystem uint64, security uint64) {

	NetworkTreasury.Ecosystem += ecosystem
	NetworkTreasury.Security += security

	record := TreasuryRecord{
		ID:     uint64(len(TreasuryHistory) + 1),
		Type:   "Deposit",
		Amount: ecosystem + security,
		Reason: "Reward Distribution",
		Block:  0,
		Time:   time.Now(),
	}

	TreasuryHistory = append(TreasuryHistory, record)
}

func GetTreasuryBalance() Treasury {
	return NetworkTreasury
}
