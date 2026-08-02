package app

import "time"

const (
	ValidatorRewardLaunch = 80
	TreasuryRewardLaunch  = 10
	SecurityRewardLaunch  = 10

	ValidatorRewardMature = 80
	TreasuryRewardMature  = 15
	SecurityRewardMature  = 5
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
	Fee        uint64
	Reward     uint64
	Time       time.Time
	Claimed    bool
}

var CurrentRewardPool RewardPool

var RewardHistory []RewardRecord
func CalculateReward(fee uint64, mature bool) RewardPool {

	var pool RewardPool

	if mature {

		pool.Validator = fee * ValidatorRewardMature / 100
		pool.Treasury = fee * TreasuryRewardMature / 100
		pool.Security = fee * SecurityRewardMature / 100

	} else {

		pool.Validator = fee * ValidatorRewardLaunch / 100
		pool.Treasury = fee * TreasuryRewardLaunch / 100
		pool.Security = fee * SecurityRewardLaunch / 100
	}

	return pool
}

func UpdateRewardPool(pool RewardPool) {

	CurrentRewardPool.Validator += pool.Validator
	CurrentRewardPool.Treasury += pool.Treasury
	CurrentRewardPool.Security += pool.Security
}

func AddRewardHistory(record RewardRecord) {

	record.ID = uint64(len(RewardHistory) + 1)

	RewardHistory = append(RewardHistory, record)
}
func DistributeReward(
	validator string,
	block uint64,
	fee uint64,
	mature bool,
) {

	pool := CalculateReward(fee, mature)

	UpdateRewardPool(pool)
        
        DepositTreasury(pool.Treasury, pool.Security)

	record := RewardRecord{
		Validator: validator,
		Block:     block,
		Fee:       fee,
		Reward:    pool.Validator,
		Time:      time.Now(),
		Claimed:   false,
	}

	AddRewardHistory(record)
}
func ClaimReward(validator string) uint64 {

	var total uint64

	for i := range RewardHistory {

		if RewardHistory[i].Validator == validator &&
			!RewardHistory[i].Claimed {

			total += RewardHistory[i].Reward

			RewardHistory[i].Claimed = true
		}
	}

        if total > 0 {
        CreditBalance(validator, total)
}
	return total
}
