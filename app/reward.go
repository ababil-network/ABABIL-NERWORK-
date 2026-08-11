package app

import (
	"errors"
	"math"
	"time"
)

const (
	ValidatorRewardLaunch = 80
	TreasuryRewardLaunch  = 10
	SecurityRewardLaunch  = 10

	ValidatorRewardMature = 80
	TreasuryRewardMature  = 15
	SecurityRewardMature  = 5
)

var (
	errRewardPoolOverflow = errors.New("reward pool balance overflow")
	errRewardHistoryFull  = errors.New("reward history ID overflow")
)

type RewardPool struct {
	Validator uint64
	Treasury  uint64
	Security  uint64
}

type RewardRecord struct {
	ID        uint64
	Validator string
	Block     uint64
	Fee       uint64
	Reward    uint64
	Time      time.Time
	Claimed   bool
}

var CurrentRewardPool RewardPool
var RewardHistory []RewardRecord

func CalculateReward(fee uint64, mature bool) RewardPool {
	var pool RewardPool

	if mature {
		pool.Validator = calculatePercentage(fee, ValidatorRewardMature)
		pool.Treasury = calculatePercentage(fee, TreasuryRewardMature)
		pool.Security = calculatePercentage(fee, SecurityRewardMature)
	} else {
		pool.Validator = calculatePercentage(fee, ValidatorRewardLaunch)
		pool.Treasury = calculatePercentage(fee, TreasuryRewardLaunch)
		pool.Security = calculatePercentage(fee, SecurityRewardLaunch)
	}

	return pool
}

// UpdateRewardPool adds a reward pool atomically.
// If any field would overflow, no field is modified.
func UpdateRewardPool(pool RewardPool) error {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	if pool.Validator > math.MaxUint64-CurrentRewardPool.Validator {
		return errRewardPoolOverflow
	}

	if pool.Treasury > math.MaxUint64-CurrentRewardPool.Treasury {
		return errRewardPoolOverflow
	}

	if pool.Security > math.MaxUint64-CurrentRewardPool.Security {
		return errRewardPoolOverflow
	}

	CurrentRewardPool.Validator += pool.Validator
	CurrentRewardPool.Treasury += pool.Treasury
	CurrentRewardPool.Security += pool.Security

	return nil
}

func AddRewardHistory(record RewardRecord) error {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	if uint64(len(RewardHistory)) == math.MaxUint64 {
		return errRewardHistoryFull
	}

	record.ID = uint64(len(RewardHistory)) + 1
	RewardHistory = append(RewardHistory, record)

	return nil
}

func DistributeReward(
	validator string,
	block uint64,
	fee uint64,
	mature bool,
) error {
	if fee == 0 {
		return nil
	}

	leader := GetLeader()
	if leader == nil {
		return nil
	}

	if leader.Address != validator {
		return nil
	}

	if leader.Jailed || !leader.Active {
		return nil
	}

	pool := CalculateReward(fee, mature)

	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	// Validate every shared-state mutation before changing anything.
	if pool.Validator > math.MaxUint64-CurrentRewardPool.Validator {
		return errRewardPoolOverflow
	}

	if pool.Treasury > math.MaxUint64-CurrentRewardPool.Treasury {
		return errRewardPoolOverflow
	}

	if pool.Security > math.MaxUint64-CurrentRewardPool.Security {
		return errRewardPoolOverflow
	}

	if pool.Treasury > math.MaxUint64-NetworkTreasury.Ecosystem {
		return errTreasuryOverflow
	}

	if pool.Security > math.MaxUint64-NetworkTreasury.Security {
		return errTreasuryOverflow
	}

	if pool.Treasury > math.MaxUint64-pool.Security {
		return errTreasuryRecordOverflow
	}

	if uint64(len(RewardHistory)) == math.MaxUint64 {
		return errRewardHistoryFull
	}

	if uint64(len(TreasuryHistory)) == math.MaxUint64 {
		return errTreasuryRecordOverflow
	}

	// All overflow checks passed. Commit the complete distribution.
	CurrentRewardPool.Validator += pool.Validator
	CurrentRewardPool.Treasury += pool.Treasury
	CurrentRewardPool.Security += pool.Security

	NetworkTreasury.Ecosystem += pool.Treasury
	NetworkTreasury.Security += pool.Security

	now := time.Now()

	TreasuryHistory = append(TreasuryHistory, TreasuryRecord{
		ID:     uint64(len(TreasuryHistory)) + 1,
		Type:   "Deposit",
		Amount: pool.Treasury + pool.Security,
		Reason: "Reward Distribution",
		Block:  block,
		Time:   now,
	})

	RewardHistory = append(RewardHistory, RewardRecord{
		ID:        uint64(len(RewardHistory)) + 1,
		Validator: validator,
		Block:     block,
		Fee:       fee,
		Reward:    pool.Validator,
		Time:      now,
		Claimed:   false,
	})

	return nil
}

func ClaimReward(validator string) uint64 {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	var total uint64

	for i := range RewardHistory {
		if RewardHistory[i].Validator != validator ||
			RewardHistory[i].Claimed {
			continue
		}

		reward := RewardHistory[i].Reward

		if reward > math.MaxUint64-total {
			return 0
		}

		total += reward
	}

	if total == 0 {
		return 0
	}

	// Do not mark rewards claimed until the balance credit succeeds.
	if GetBalance(validator) > math.MaxUint64-total {
		return 0
	}

	if err := CreditBalance(validator, total); err != nil {
		return 0
	}

	for i := range RewardHistory {
		if RewardHistory[i].Validator == validator &&
			!RewardHistory[i].Claimed {
			RewardHistory[i].Claimed = true
		}
	}

	return total
}

func GetRewardPool() RewardPool {
	rewardStateMu.Lock()
	defer rewardStateMu.Unlock()

	return CurrentRewardPool
}
