package app

import (
	"errors"
	"testing"
)

func resetLeaderRotationBoundaryStateForTest() {
	validatorStateMu.Lock()
	oldValidators := Validators
	oldLeaderIndex := LeaderIndex
	Validators = nil
	LeaderIndex = 0
	validatorStateMu.Unlock()

	blockchainMu.Lock()
	oldBlockchain := Blockchain
	Blockchain = nil
	blockchainMu.Unlock()

	_ = oldValidators
	_ = oldLeaderIndex
	_ = oldBlockchain
}

func setupLeaderRotationBoundaryValidatorsForTest(t *testing.T) {
	t.Helper()

	validatorStateMu.Lock()
	Validators = []Validator{
		{
			ID:      1,
			Address: "validator-1",
			Power:   100,
			Active:  true,
			Jailed:  false,
		},
		{
			ID:      2,
			Address: "validator-2",
			Power:   100,
			Active:  true,
			Jailed:  false,
		},
		{
			ID:      3,
			Address: "validator-3",
			Power:   100,
			Active:  true,
			Jailed:  false,
		},
	}
	LeaderIndex = 0
	validatorStateMu.Unlock()

	t.Cleanup(func() {
		validatorStateMu.Lock()
		Validators = nil
		LeaderIndex = 0
		validatorStateMu.Unlock()
	})
}

func TestLeaderRotationBoundary_CurrentLeaderBaseline(t *testing.T) {
	setupLeaderRotationBoundaryValidatorsForTest(t)

	leader := GetLeader()
	if leader == nil {
		t.Fatal("expected current leader")
	}

	if leader.Address != "validator-1" {
		t.Fatalf("expected validator-1, got %s", leader.Address)
	}

	leaderIndex := func() int {
		validatorStateMu.RLock()
		defer validatorStateMu.RUnlock()
		return LeaderIndex
	}()

	if leaderIndex != 0 {
		t.Fatalf("expected LeaderIndex=0, got %d", leaderIndex)
	}
}

func TestLeaderRotationBoundary_RotateLeaderAdvancesExactlyOnce(t *testing.T) {
	setupLeaderRotationBoundaryValidatorsForTest(t)

	first := RotateLeader()
	if first == nil {
		t.Fatal("first rotation returned nil")
	}

	if first.Address != "validator-2" {
		t.Fatalf("expected validator-2 after first rotation, got %s", first.Address)
	}

	validatorStateMu.RLock()
	firstIndex := LeaderIndex
	validatorStateMu.RUnlock()

	if firstIndex != 1 {
		t.Fatalf("expected LeaderIndex=1 after first rotation, got %d", firstIndex)
	}

	second := RotateLeader()
	if second == nil {
		t.Fatal("second rotation returned nil")
	}

	if second.Address != "validator-3" {
		t.Fatalf("expected validator-3 after second rotation, got %s", second.Address)
	}
}

func TestLeaderRotationBoundary_FailedCommitDoesNotRotate(t *testing.T) {
	setupLeaderRotationBoundaryValidatorsForTest(t)

	blockchainMu.Lock()
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
	}
	blockchainMu.Unlock()

	before := GetLeader()
	if before == nil {
		t.Fatal("expected leader before failed commit")
	}

	block := Block{
		Height:       2,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "INVALID_GAP_BLOCK",
	}

	err := CommitBlock(block)
	if !errors.Is(err, ErrBlockHeightGap) {
		t.Fatalf("expected ErrBlockHeightGap, got %v", err)
	}

	after := GetLeader()
	if after == nil {
		t.Fatal("expected leader after failed commit")
	}

	if after.Address != before.Address {
		t.Fatalf(
			"leader rotated after failed commit: before=%s after=%s",
			before.Address,
			after.Address,
		)
	}
}

func TestLeaderRotationBoundary_DuplicateCommitDoesNotRotate(t *testing.T) {
	setupLeaderRotationBoundaryValidatorsForTest(t)

	blockchainMu.Lock()
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	before := GetLeader()
	if before == nil {
		t.Fatal("expected leader before duplicate commit")
	}

	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "BLOCK_1",
	}

	err := CommitBlock(block)
	if !errors.Is(err, ErrBlockAlreadyCommitted) {
		t.Fatalf("expected ErrBlockAlreadyCommitted, got %v", err)
	}

	after := GetLeader()
	if after == nil {
		t.Fatal("expected leader after duplicate commit")
	}

	if after.Address != before.Address {
		t.Fatalf(
			"leader rotated after duplicate commit: before=%s after=%s",
			before.Address,
			after.Address,
		)
	}
}

func TestLeaderRotationBoundary_ConflictingCommitDoesNotRotate(t *testing.T) {
	setupLeaderRotationBoundaryValidatorsForTest(t)

	blockchainMu.Lock()
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	before := GetLeader()
	if before == nil {
		t.Fatal("expected leader before conflicting commit")
	}

	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "ATTACKER_BLOCK",
	}

	err := CommitBlock(block)
	if !errors.Is(err, ErrBlockConflict) {
		t.Fatalf("expected ErrBlockConflict, got %v", err)
	}

	after := GetLeader()
	if after == nil {
		t.Fatal("expected leader after conflicting commit")
	}

	if after.Address != before.Address {
		t.Fatalf(
			"leader rotated after conflicting commit: before=%s after=%s",
			before.Address,
			after.Address,
		)
	}
}
