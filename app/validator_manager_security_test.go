package app

import (
	"testing"
	"time"
)

func resetValidatorRegistrationStateForTest() {
	Validators = nil
	ValidatorCollaterals = nil
	LeaderIndex = 0
	WalletBalances = nil
	walletBalanceIndex = nil
	NodeReferencePrice.Reset()
}

func setupValidatorRegistrationTestEnvironment(t *testing.T) {
	t.Helper()

	NodeReferencePrice.Reset()

	if err := NodeReferencePrice.AddObservation(
		1,
		time.Now().UTC(),
	); err != nil {
		t.Fatalf("failed to initialize reference price: %v", err)
	}
}

func TestRegisterValidatorAtomicSuccess(t *testing.T) {
	resetValidatorRegistrationStateForTest()
	defer resetValidatorRegistrationStateForTest()

	setupValidatorRegistrationTestEnvironment(t)

	address := "0x1111111111111111111111111111111111111111"

	WalletBalances = []WalletBalance{
		{
			Address: address,
			Balance: 100,
		},
	}
	walletBalanceIndex = nil

	if err := RegisterValidator(address, MinimumValidatorPower, 5); err != nil {
		t.Fatalf("validator registration failed: %v", err)
	}

	if len(Validators) != 1 {
		t.Fatalf("expected 1 validator, got %d", len(Validators))
	}

	if Validators[0].Address != address {
		t.Fatalf("unexpected validator address: %s", Validators[0].Address)
	}

	if Validators[0].Power != MinimumValidatorPower {
		t.Fatalf(
			"unexpected validator power: got %d want %d",
			Validators[0].Power,
			MinimumValidatorPower,
		)
	}

	if Validators[0].Commission != 5 {
		t.Fatalf(
			"unexpected commission: got %d want 5",
			Validators[0].Commission,
		)
	}

	collateral, err := GetValidatorCollateral(address)
	if err != nil {
		t.Fatalf("expected collateral record: %v", err)
	}

	if collateral.Amount != 0 {
		t.Fatalf(
			"slot 1 must have zero collateral, got %d",
			collateral.Amount,
		)
	}

	if !collateral.Locked {
		t.Fatal("validator collateral record must remain locked")
	}

	if got := GetBalance(address); got != 100 {
		t.Fatalf(
			"balance changed unexpectedly: got %d want 100",
			got,
		)
	}
}

func TestRegisterValidatorInsufficientBalanceIsAtomic(t *testing.T) {
	resetValidatorRegistrationStateForTest()
	defer resetValidatorRegistrationStateForTest()

	setupValidatorRegistrationTestEnvironment(t)

	target := "0x2222222222222222222222222222222222222222"

	// Create exactly 10 existing validator slots.
	Validators = make([]Validator, 10)

	for i := range Validators {
		Validators[i] = Validator{
			ID:         uint64(i + 1),
			Address:    "0x1111111111111111111111111111111111111111",
			Power:      MinimumValidatorPower,
			Commission: 5,
			Active:     true,
		}

		// Make every address unique.
		Validators[i].Address = "0x" + string(rune('a'+i)) +
			"111111111111111111111111111111111111"
	}

	WalletBalances = []WalletBalance{
		{
			Address: target,
			Balance: 149,
		},
	}
	walletBalanceIndex = nil

	beforeValidators := len(Validators)

	err := RegisterValidator(
		target,
		MinimumValidatorPower,
		5,
	)

	if err != ErrInsufficientFunds {
		t.Fatalf(
			"expected insufficient funds, got %v",
			err,
		)
	}

	if len(Validators) != beforeValidators {
		t.Fatalf(
			"validator state mutated after failed registration: got %d want %d",
			len(Validators),
			beforeValidators,
		)
	}

	if got := GetBalance(target); got != 149 {
		t.Fatalf(
			"balance changed after failed registration: got %d want 149",
			got,
		)
	}

	if _, err := GetValidatorCollateral(target); err != ErrValidatorCollateralNotFound {
		t.Fatalf(
			"unexpected collateral state after failed registration: %v",
			err,
		)
	}
}

func TestRegisterValidatorDuplicateIsAtomic(t *testing.T) {
	resetValidatorRegistrationStateForTest()
	defer resetValidatorRegistrationStateForTest()

	setupValidatorRegistrationTestEnvironment(t)

	address := "0x3333333333333333333333333333333333333333"

	WalletBalances = []WalletBalance{
		{
			Address: address,
			Balance: 1000,
		},
	}
	walletBalanceIndex = nil

	if err := RegisterValidator(
		address,
		MinimumValidatorPower,
		5,
	); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	beforeBalance := GetBalance(address)
	beforeValidators := len(Validators)

	err := RegisterValidator(
		address,
		MinimumValidatorPower,
		5,
	)

	if err == nil {
		t.Fatal("expected duplicate validator registration to fail")
	}

	if len(Validators) != beforeValidators {
		t.Fatal("validator count changed after duplicate registration")
	}

	if got := GetBalance(address); got != beforeBalance {
		t.Fatalf(
			"balance changed after duplicate registration: got %d want %d",
			got,
			beforeBalance,
		)
	}
}

func TestRegisterValidatorInvalidInputIsAtomic(t *testing.T) {
	resetValidatorRegistrationStateForTest()
	defer resetValidatorRegistrationStateForTest()

	setupValidatorRegistrationTestEnvironment(t)

	address := "0x4444444444444444444444444444444444444444"

	WalletBalances = []WalletBalance{
		{
			Address: address,
			Balance: 1000,
		},
	}
	walletBalanceIndex = nil

	beforeBalance := GetBalance(address)

	err := RegisterValidator(
		address,
		MinimumValidatorPower-1,
		5,
	)

	if err == nil {
		t.Fatal("expected invalid validator power to be rejected")
	}

	if len(Validators) != 0 {
		t.Fatal("validator state mutated after invalid registration")
	}

	if got := GetBalance(address); got != beforeBalance {
		t.Fatalf(
			"balance changed after invalid registration: got %d want %d",
			got,
			beforeBalance,
		)
	}

	if _, err := GetValidatorCollateral(address); err != ErrValidatorCollateralNotFound {
		t.Fatalf(
			"unexpected collateral after invalid registration: %v",
			err,
		)
	}
}
