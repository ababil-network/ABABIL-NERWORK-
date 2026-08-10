package app

import (
	"math"
	"testing"
)

func TestUpdateRewardPoolDoesNotOverflow(t *testing.T) {
	original := CurrentRewardPool
	defer func() {
		CurrentRewardPool = original
	}()

	CurrentRewardPool = RewardPool{
		Validator: math.MaxUint64,
		Treasury:  math.MaxUint64,
		Security:  math.MaxUint64,
	}

	UpdateRewardPool(RewardPool{
		Validator: 1,
		Treasury:  1,
		Security:  1,
	})

	if CurrentRewardPool.Validator != math.MaxUint64 {
		t.Fatalf("validator reward pool wrapped: got %d", CurrentRewardPool.Validator)
	}

	if CurrentRewardPool.Treasury != math.MaxUint64 {
		t.Fatalf("treasury reward pool wrapped: got %d", CurrentRewardPool.Treasury)
	}

	if CurrentRewardPool.Security != math.MaxUint64 {
		t.Fatalf("security reward pool wrapped: got %d", CurrentRewardPool.Security)
	}
}

func TestDepositTreasuryDoesNotOverflow(t *testing.T) {
	originalTreasury := NetworkTreasury
	originalHistory := TreasuryHistory

	defer func() {
		NetworkTreasury = originalTreasury
		TreasuryHistory = originalHistory
	}()

	NetworkTreasury = Treasury{
		Ecosystem: math.MaxUint64,
		Security:  math.MaxUint64,
	}
	TreasuryHistory = nil

	DepositTreasury(1, 1)

	if NetworkTreasury.Ecosystem != math.MaxUint64 {
		t.Fatalf("ecosystem treasury wrapped: got %d", NetworkTreasury.Ecosystem)
	}

	if NetworkTreasury.Security != math.MaxUint64 {
		t.Fatalf("security treasury wrapped: got %d", NetworkTreasury.Security)
	}
}

func TestCalculateRewardConservesFee(t *testing.T) {
	fees := []uint64{
		0,
		1,
		99,
		100,
		101,
		1_000_000,
		math.MaxUint64,
	}

	for _, fee := range fees {
		for _, mature := range []bool{false, true} {
			pool := CalculateReward(fee, mature)

			total := pool.Validator + pool.Treasury + pool.Security

			if total > fee {
				t.Fatalf(
					"reward exceeds fee: fee=%d mature=%v validator=%d treasury=%d security=%d total=%d",
					fee,
					mature,
					pool.Validator,
					pool.Treasury,
					pool.Security,
					total,
				)
			}
		}
	}
}
