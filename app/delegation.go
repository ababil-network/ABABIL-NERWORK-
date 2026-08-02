package app

import "time"

type Delegation struct {
	ID          uint64
	Delegator   string
	Validator   string
	Amount      uint64
	Active      bool
	CreatedAt   time.Time
}

type DelegationHistory struct {
	ID          uint64
	Delegator   string
	Validator   string
	Action      string
	Amount      uint64
	Time        time.Time
}

var Delegations []Delegation
var DelegationRecords []DelegationHistory
