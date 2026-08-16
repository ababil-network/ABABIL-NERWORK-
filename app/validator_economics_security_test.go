package app

import "testing"

func TestValidatorDepositSchedule(t *testing.T) {
	tests := []struct {
		slot uint64
		want uint64
	}{
		{1, 0},
		{10, 0},
		{11, 150},
		{20, 150},
		{21, 300},
		{30, 300},
		{31, 450},
		{40, 450},
		{41, 600},
		{50, 600},
		{51, 700},
		{60, 700},
		{61, 800},
		{70, 800},
		{71, 900},
		{80, 900},
		{81, 1000},
		{90, 1000},
		{91, 1500},
		{100, 1500},
		{1000, 1500},
	}

	for _, tt := range tests {
		got, err := ValidatorDepositUSD(tt.slot)
		if err != nil {
			t.Fatalf("slot %d: unexpected error: %v", tt.slot, err)
		}

		if got != tt.want {
			t.Fatalf(
				"slot %d: expected $%d deposit, got $%d",
				tt.slot,
				tt.want,
				got,
			)
		}
	}
}

func TestValidatorDepositRejectsInvalidSlot(t *testing.T) {
	if _, err := ValidatorDepositUSD(0); err == nil {
		t.Fatal("expected slot zero to be rejected")
	}
}

func TestValidatorDepositConversion(t *testing.T) {
	// Example reference price:
	// 1 USD = 2,000,000 micro-ABABIL.
	price := uint64(2_000_000)

	got, err := ValidatorDepositMicroABABIL(11, price)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := uint64(150) * price

	if got != want {
		t.Fatalf("expected %d micro-ABABIL, got %d", want, got)
	}
}

func TestValidatorDepositFreeGenesisSlots(t *testing.T) {
	price := uint64(2_000_000)

	for slot := uint64(1); slot <= 10; slot++ {
		got, err := ValidatorDepositMicroABABIL(slot, price)
		if err != nil {
			t.Fatalf("slot %d: unexpected error: %v", slot, err)
		}

		if got != 0 {
			t.Fatalf(
				"slot %d: genesis validator must require zero deposit, got %d",
				slot,
				got,
			)
		}
	}
}

func TestValidatorDepositOverflowProtection(t *testing.T) {
	_, err := ValidatorDepositMicroABABIL(
		91,
		^uint64(0),
	)

	if err == nil {
		t.Fatal("expected overflow protection")
	}
}

func TestValidatorDepositInsufficient(t *testing.T) {
	price := uint64(2_000_000)

	required, err := ValidatorDepositMicroABABIL(91, price)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ValidateValidatorDeposit(
		91,
		required-1,
		price,
	); err != ErrValidatorDepositInsufficient {
		t.Fatalf(
			"expected ErrValidatorDepositInsufficient, got %v",
			err,
		)
	}
}

func TestValidatorDepositExactAmount(t *testing.T) {
	price := uint64(2_000_000)

	required, err := ValidatorDepositMicroABABIL(91, price)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := ValidateValidatorDeposit(
		91,
		required,
		price,
	); err != nil {
		t.Fatalf("exact collateral should be accepted: %v", err)
	}
}
