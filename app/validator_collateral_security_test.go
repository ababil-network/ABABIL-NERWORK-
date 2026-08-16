package app

import (
	"testing"
	"time"
)

func resetValidatorCollateralStateForTest() {
	validatorCollateralMu.Lock()
	defer validatorCollateralMu.Unlock()

	ValidatorCollaterals = nil
}

func setupValidatorCollateralReferencePriceForTest(t *testing.T) {
	t.Helper()

	NodeReferencePrice.Reset()

	// 1 USD = 1 micro-ABABIL for deterministic security-collateral tests.
	//
	// Therefore:
	// slot 11 = $150 = 150 micro-ABABIL
	// slot 21 = $300 = 300 micro-ABABIL
	price := uint64(1)

	if err := NodeReferencePrice.AddObservation(
		price,
		time.Now().UTC(),
	); err != nil {
		t.Fatalf("failed to initialize reference price: %v", err)
	}
}

func TestLockValidatorCollateral(t *testing.T) {
	resetValidatorCollateralStateForTest()
	setupValidatorCollateralReferencePriceForTest(t)

	if err := LockValidatorCollateral("validator-1", 11, 150); err != nil {
		t.Fatalf("unexpected collateral lock error: %v", err)
	}

	got, err := GetValidatorCollateral("validator-1")
	if err != nil {
		t.Fatalf("failed to retrieve collateral: %v", err)
	}

	if got.Amount != 150 {
		t.Fatalf("expected collateral 150, got %d", got.Amount)
	}

	if !got.Locked {
		t.Fatal("validator collateral must remain locked")
	}
}

func TestValidatorCollateralCannotBeDuplicated(t *testing.T) {
	resetValidatorCollateralStateForTest()
	setupValidatorCollateralReferencePriceForTest(t)

	if err := LockValidatorCollateral("validator-1", 11, 150); err != nil {
		t.Fatalf("first lock failed: %v", err)
	}

	if err := LockValidatorCollateral("validator-1", 11, 150); err != ErrValidatorCollateralExists {
		t.Fatalf("expected duplicate collateral error, got %v", err)
	}
}

func TestValidatorCollateralCannotBeReleasedNormally(t *testing.T) {
	resetValidatorCollateralStateForTest()
	setupValidatorCollateralReferencePriceForTest(t)

	if err := LockValidatorCollateral("validator-1", 11, 150); err != nil {
		t.Fatalf("lock failed: %v", err)
	}

	if err := ReleaseValidatorCollateral("validator-1"); err != ErrValidatorCollateralLocked {
		t.Fatalf("expected locked error, got %v", err)
	}

	if !IsValidatorCollateralLocked("validator-1") {
		t.Fatal("collateral was unexpectedly unlocked")
	}
}

func TestValidatorCollateralRequiresEnoughAmount(t *testing.T) {
	resetValidatorCollateralStateForTest()
	setupValidatorCollateralReferencePriceForTest(t)

	if err := LockValidatorCollateral("validator-1", 11, 149); err != ErrValidatorCollateralInsufficient {
		t.Fatalf("expected insufficient collateral error, got %v", err)
	}
}

func TestGenesisValidatorHasZeroRequiredCollateral(t *testing.T) {
	resetValidatorCollateralStateForTest()
	setupValidatorCollateralReferencePriceForTest(t)

	if err := LockValidatorCollateral("validator-1", 1, 0); err != nil {
		t.Fatalf("genesis zero-collateral validator should be allowed: %v", err)
	}

	got, err := GetValidatorCollateral("validator-1")
	if err != nil {
		t.Fatalf("failed to read genesis collateral: %v", err)
	}

	if got.Amount != 0 {
		t.Fatalf("expected zero collateral, got %d", got.Amount)
	}

	if !got.Locked {
		t.Fatal("genesis collateral record must still be marked locked")
	}
}

func TestValidatorCollateralInvalidInput(t *testing.T) {
	resetValidatorCollateralStateForTest()

	if err := LockValidatorCollateral("", 11, 150); err != ErrValidatorCollateralInvalid {
		t.Fatalf("expected invalid validator error, got %v", err)
	}

	if err := LockValidatorCollateral("validator-1", 0, 150); err != ErrValidatorCollateralInvalid {
		t.Fatalf("expected invalid slot error, got %v", err)
	}
}
