package app

import "time"

type CommissionRecord struct {
	ID          uint64
	Validator   string
	Rate        uint8
	Amount      uint64
	Claimed     bool
	CreatedAt   time.Time
	ClaimedAt   time.Time
}

var CommissionHistory []CommissionRecord
