package app

import "testing"

func resetLeaderSemanticsStateForTest() {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	Validators = nil
	LeaderIndex = 0
}

func setupLeaderSemanticsValidatorsForTest(t *testing.T) {
	t.Helper()

	resetLeaderSemanticsStateForTest()
	t.Cleanup(resetLeaderSemanticsStateForTest)

	for i := 0; i < 3; i++ {
		AddValidator(
			"0x"+formatLeaderTestAddress(i+1),
			MinimumValidatorPower,
		)
	}
}

func formatLeaderTestAddress(n int) string {
	const hex = "0123456789abcdef"

	out := make([]byte, 40)
	for i := range out {
		out[i] = hex[(n+i)%len(hex)]
	}
	return string(out)
}

// E4-C invariant:
// RotateLeader() and GetLeader() must refer to the same
// authoritative current leader.
func TestLeaderSemantics_RotationAndGetLeaderAgree(t *testing.T) {
	setupLeaderSemanticsValidatorsForTest(t)

	first := GetLeader()
	if first == nil {
		t.Fatal("GetLeader returned nil with active validators")
	}

	rotated := RotateLeader()
	if rotated == nil {
		t.Fatal("RotateLeader returned nil with active validators")
	}

	current := GetLeader()
	if current == nil {
		t.Fatal("GetLeader returned nil after rotation")
	}

	if current.Address != rotated.Address {
		t.Fatalf(
			"leader semantic mismatch: RotateLeader=%s, GetLeader=%s",
			rotated.Address,
			current.Address,
		)
	}

	if current.Address == first.Address {
		t.Fatalf(
			"leader did not rotate: first=%s current=%s",
			first.Address,
			current.Address,
		)
	}
}

// E4-C invariant:
// A jailed/inactive validator must never remain the effective leader.
func TestLeaderSemantics_JailedLeaderIsSkipped(t *testing.T) {
	setupLeaderSemanticsValidatorsForTest(t)

	initial := GetLeader()
	if initial == nil {
		t.Fatal("expected initial leader")
	}

	initialAddress := initial.Address

	if !JailValidator(initialAddress) {
		t.Fatalf("failed to jail current leader %s", initialAddress)
	}

	current := GetLeader()
	if current == nil {
		t.Fatal("GetLeader returned nil although active validators remain")
	}

	if current.Address == initialAddress {
		t.Fatalf(
			"jailed validator remained effective leader: %s",
			initialAddress,
		)
	}

	if current.Jailed || !current.Active {
		t.Fatalf(
			"GetLeader returned invalid leader: address=%s active=%v jailed=%v",
			current.Address,
			current.Active,
			current.Jailed,
		)
	}
}

// E4-C invariant:
// Repeated rotations must always return an eligible validator and
// LeaderIndex must remain valid.
func TestLeaderSemantics_RepeatedRotationMaintainsInvariant(t *testing.T) {
	setupLeaderSemanticsValidatorsForTest(t)

	const rotations = 100

	for i := 0; i < rotations; i++ {
		leader := RotateLeader()
		if leader == nil {
			t.Fatalf("rotation %d returned nil", i)
		}

		validatorStateMu.RLock()

		if LeaderIndex < 0 || LeaderIndex >= len(Validators) {
			validatorStateMu.RUnlock()

			t.Fatalf(
				"rotation %d produced invalid LeaderIndex=%d validators=%d",
				i,
				LeaderIndex,
				len(Validators),
			)
		}

		indexed := Validators[LeaderIndex]

		validatorStateMu.RUnlock()

		if indexed.Address != leader.Address {
			t.Fatalf(
				"rotation %d mismatch: LeaderIndex points to %s, RotateLeader returned %s",
				i,
				indexed.Address,
				leader.Address,
			)
		}

		if leader.Jailed || !leader.Active {
			t.Fatalf(
				"rotation %d selected invalid validator: %s active=%v jailed=%v",
				i,
				leader.Address,
				leader.Active,
				leader.Jailed,
			)
		}

		current := GetLeader()
		if current == nil {
			t.Fatalf("rotation %d: GetLeader returned nil", i)
		}

		if current.Address != leader.Address {
			t.Fatalf(
				"rotation %d semantic mismatch: rotated=%s current=%s",
				i,
				leader.Address,
				current.Address,
			)
		}
	}
}
