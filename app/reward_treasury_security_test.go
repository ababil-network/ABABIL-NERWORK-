package app

import (
	"math"
	"sync"
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

func TestDepositTreasuryConcurrent(t *testing.T) {
	originalTreasury := NetworkTreasury
	originalHistory := TreasuryHistory

	defer func() {
		NetworkTreasury = originalTreasury
		TreasuryHistory = originalHistory
	}()

	NetworkTreasury = Treasury{}
	TreasuryHistory = nil

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			if err := DepositTreasury(1, 1); err != nil {
				t.Errorf("DepositTreasury failed: %v", err)
			}
		}()
	}

	wg.Wait()

	if NetworkTreasury.Ecosystem != workers {
		t.Fatalf(
			"unexpected ecosystem treasury: got %d want %d",
			NetworkTreasury.Ecosystem,
			workers,
		)
	}

	if NetworkTreasury.Security != workers {
		t.Fatalf(
			"unexpected security treasury: got %d want %d",
			NetworkTreasury.Security,
			workers,
		)
	}

	if len(TreasuryHistory) != workers {
		t.Fatalf(
			"unexpected treasury history length: got %d want %d",
			len(TreasuryHistory),
			workers,
		)
	}

	for i, record := range TreasuryHistory {
		wantID := uint64(i + 1)

		if record.ID != wantID {
			t.Fatalf(
				"duplicate or invalid treasury history ID at index %d: got %d want %d",
				i,
				record.ID,
				wantID,
			)
		}
	}
}

func TestClaimRewardConcurrentDoubleClaim(t *testing.T) {
	originalHistory := RewardHistory
	originalBalances := WalletBalances
	originalIndex := walletBalanceIndex

	defer func() {
		RewardHistory = originalHistory
		WalletBalances = originalBalances
		walletBalanceIndex = originalIndex
	}()

	const validator = "0x7777777777777777777777777777777777777777"
	const reward = uint64(100)

	RewardHistory = []RewardRecord{
		{
			ID:        1,
			Validator: validator,
			Block:     1,
			Fee:       1000,
			Reward:    reward,
			Claimed:   false,
		},
	}

	WalletBalances = nil
	walletBalanceIndex = make(map[string]int)

	if err := CreditBalance(validator, 0); err != nil {
		t.Fatalf("failed to initialize validator balance: %v", err)
	}

	const workers = 100

	var wg sync.WaitGroup
	results := make(chan uint64, workers)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			results <- ClaimReward(validator)
		}()
	}

	wg.Wait()
	close(results)

	var totalClaimed uint64
	successes := 0

	for amount := range results {
		if amount > 0 {
			successes++
			totalClaimed += amount
		}
	}

	if successes != 1 {
		t.Fatalf(
			"expected exactly one successful claim, got %d",
			successes,
		)
	}

	if totalClaimed != reward {
		t.Fatalf(
			"unexpected total claimed amount: got %d want %d",
			totalClaimed,
			reward,
		)
	}

	if got := GetBalance(validator); got != reward {
		t.Fatalf(
			"unexpected validator balance: got %d want %d",
			got,
			reward,
		)
	}

	if !RewardHistory[0].Claimed {
		t.Fatal("reward was not marked as claimed")
	}
}

func TestClaimRewardCreditFailureDoesNotMarkClaimed(t *testing.T) {
	originalHistory := RewardHistory
	originalBalances := WalletBalances
	originalIndex := walletBalanceIndex

	defer func() {
		RewardHistory = originalHistory
		WalletBalances = originalBalances
		walletBalanceIndex = originalIndex
	}()

	const validator = "0x8888888888888888888888888888888888888888"
	const reward = uint64(100)

	RewardHistory = []RewardRecord{
		{
			ID:        1,
			Validator: validator,
			Block:     1,
			Fee:       1000,
			Reward:    reward,
			Claimed:   false,
		},
	}

	WalletBalances = []WalletBalance{
		{
			Address: validator,
			Balance: math.MaxUint64 - reward + 1,
		},
	}
	walletBalanceIndex = nil

	claimed := ClaimReward(validator)

	if claimed != 0 {
		t.Fatalf("expected failed claim to return 0, got %d", claimed)
	}

	if RewardHistory[0].Claimed {
		t.Fatal("reward was marked claimed after CreditBalance failure")
	}

	if got := GetBalance(validator); got != math.MaxUint64-reward+1 {
		t.Fatalf(
			"balance changed after failed reward claim: got %d want %d",
			got,
			uint64(math.MaxUint64-reward+1),
		)
	}
}

func TestDistributeRewardConcurrentAtomicity(t *testing.T) {
	originalValidators := Validators
	originalRewardPool := CurrentRewardPool
	originalRewardHistory := RewardHistory
	originalTreasury := NetworkTreasury
	originalTreasuryHistory := TreasuryHistory

	defer func() {
		Validators = originalValidators
		CurrentRewardPool = originalRewardPool
		RewardHistory = originalRewardHistory
		NetworkTreasury = originalTreasury
		TreasuryHistory = originalTreasuryHistory
	}()

	const (
		validator = "0x9999999999999999999999999999999999999999"
		workers   = 100
		fee       = uint64(1000)
	)

	Validators = []Validator{
		{
			ID:           1,
			Address:      validator,
			ConsensusKey: "",
			Power:        100,
			Commission:   5,
			Active:       true,
			Jailed:       false,
			Genesis:      true,
		},
	}

	CurrentRewardPool = RewardPool{}
	RewardHistory = nil
	NetworkTreasury = Treasury{}
	TreasuryHistory = nil

	var wg sync.WaitGroup
	errorsCh := make(chan error, workers)

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		block := uint64(i + 1)

		go func(block uint64) {
			defer wg.Done()

			if err := DistributeReward(
				validator,
				block,
				fee,
				false,
			); err != nil {
				errorsCh <- err
			}
		}(block)
	}

	wg.Wait()
	close(errorsCh)

	for err := range errorsCh {
		t.Fatalf("concurrent reward distribution failed: %v", err)
	}

	pool := CalculateReward(fee, false)

	wantValidator := pool.Validator * workers
	wantTreasury := pool.Treasury * workers
	wantSecurity := pool.Security * workers

	if CurrentRewardPool.Validator != wantValidator {
		t.Fatalf(
			"unexpected validator reward pool: got %d want %d",
			CurrentRewardPool.Validator,
			wantValidator,
		)
	}

	if CurrentRewardPool.Treasury != wantTreasury {
		t.Fatalf(
			"unexpected treasury reward pool: got %d want %d",
			CurrentRewardPool.Treasury,
			wantTreasury,
		)
	}

	if CurrentRewardPool.Security != wantSecurity {
		t.Fatalf(
			"unexpected security reward pool: got %d want %d",
			CurrentRewardPool.Security,
			wantSecurity,
		)
	}

	if NetworkTreasury.Ecosystem != wantTreasury {
		t.Fatalf(
			"unexpected ecosystem treasury: got %d want %d",
			NetworkTreasury.Ecosystem,
			wantTreasury,
		)
	}

	if NetworkTreasury.Security != wantSecurity {
		t.Fatalf(
			"unexpected security treasury: got %d want %d",
			NetworkTreasury.Security,
			wantSecurity,
		)
	}

	if len(RewardHistory) != workers {
		t.Fatalf(
			"unexpected reward history length: got %d want %d",
			len(RewardHistory),
			workers,
		)
	}

	if len(TreasuryHistory) != workers {
		t.Fatalf(
			"unexpected treasury history length: got %d want %d",
			len(TreasuryHistory),
			workers,
		)
	}

	for i, record := range RewardHistory {
		wantID := uint64(i + 1)

		if record.ID != wantID {
			t.Fatalf(
				"invalid reward history ID at index %d: got %d want %d",
				i,
				record.ID,
				wantID,
			)
		}

		if record.Validator != validator {
			t.Fatalf(
				"unexpected validator at reward history index %d: got %s want %s",
				i,
				record.Validator,
				validator,
			)
		}
	}

	for i, record := range TreasuryHistory {
		wantID := uint64(i + 1)

		if record.ID != wantID {
			t.Fatalf(
				"invalid treasury history ID at index %d: got %d want %d",
				i,
				record.ID,
				wantID,
			)
		}
	}
}
