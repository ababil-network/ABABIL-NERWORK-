package app

import (
	"math"
	"testing"
)

func TestCalculateRewardDoesNotOverflow(t *testing.T) {
	fee := uint64(math.MaxUint64)

	pool := CalculateReward(fee, false)

	if pool.Validator > fee {
		t.Fatalf("validator reward exceeds fee: %d > %d", pool.Validator, fee)
	}

	if pool.Treasury > fee {
		t.Fatalf("treasury reward exceeds fee: %d > %d", pool.Treasury, fee)
	}

	if pool.Security > fee {
		t.Fatalf("security reward exceeds fee: %d > %d", pool.Security, fee)
	}

	if pool.Validator > math.MaxUint64-pool.Treasury {
		t.Fatal("reward pool total overflow")
	}

	total := pool.Validator + pool.Treasury

	if total > math.MaxUint64-pool.Security {
		t.Fatal("reward pool total overflow")
	}
}

func TestCalculateRewardNormalFee(t *testing.T) {
	fee := uint64(100)

	pool := CalculateReward(fee, false)

	if pool.Validator != 80 {
		t.Fatalf("validator reward: got %d want 80", pool.Validator)
	}

	if pool.Treasury != 10 {
		t.Fatalf("treasury reward: got %d want 10", pool.Treasury)
	}

	if pool.Security != 10 {
		t.Fatalf("security reward: got %d want 10", pool.Security)
	}
}

func TestCalculateRewardMatureFee(t *testing.T) {
	fee := uint64(100)

	pool := CalculateReward(fee, true)

	if pool.Validator != 80 {
		t.Fatalf("validator reward: got %d want 80", pool.Validator)
	}

	if pool.Treasury != 15 {
		t.Fatalf("treasury reward: got %d want 15", pool.Treasury)
	}

	if pool.Security != 5 {
		t.Fatalf("security reward: got %d want 5", pool.Security)
	}
}
