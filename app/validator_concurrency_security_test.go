package app

import (
	"fmt"
	"sync"
	"testing"
)

func resetValidatorConcurrencyStateForTest() {
	validatorStateMu.Lock()
	defer validatorStateMu.Unlock()

	Validators = nil
	LeaderIndex = 0
}

// Test concurrent AddValidator calls with unique addresses.
// The final validator count must never exceed the number of unique inputs.
func TestConcurrentAddValidatorUniqueAddresses(t *testing.T) {
	resetValidatorConcurrencyStateForTest()
	defer resetValidatorConcurrencyStateForTest()

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			AddValidator(
				fmt.Sprintf("0x%040x", i+1),
				MinimumValidatorPower,
			)
		}()
	}

	wg.Wait()

	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	if len(Validators) != workers {
		t.Fatalf("expected %d validators, got %d", workers, len(Validators))
	}

	seen := make(map[string]bool, len(Validators))

	for _, v := range Validators {
		if seen[v.Address] {
			t.Fatalf("duplicate validator detected: %s", v.Address)
		}

		seen[v.Address] = true
	}
}

// Test concurrent duplicate registration through AddValidator.
// Only one validator should survive for the same address.
func TestConcurrentAddValidatorDuplicateAddress(t *testing.T) {
	resetValidatorConcurrencyStateForTest()
	defer resetValidatorConcurrencyStateForTest()

	const workers = 100
	const address = "0x1111111111111111111111111111111111111111"

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			AddValidator(address, MinimumValidatorPower)
		}()
	}

	wg.Wait()

	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	if len(Validators) != 1 {
		t.Fatalf("expected exactly 1 validator, got %d", len(Validators))
	}

	if Validators[0].Address != address {
		t.Fatalf("unexpected validator address: %s", Validators[0].Address)
	}
}

// Test concurrent leader rotation.
// This verifies that LeaderIndex remains within bounds and that
// RotateLeader does not corrupt validator state.
func TestConcurrentRotateLeader(t *testing.T) {
	resetValidatorConcurrencyStateForTest()
	defer resetValidatorConcurrencyStateForTest()

	const validators = 10
	const rotations = 1000

	for i := 0; i < validators; i++ {
		AddValidator(
			fmt.Sprintf("0x%040x", i+1),
			MinimumValidatorPower,
		)
	}

	var wg sync.WaitGroup
	wg.Add(rotations)

	for i := 0; i < rotations; i++ {
		go func() {
			defer wg.Done()

			leader := RotateLeader()
			if leader == nil {
				t.Errorf("RotateLeader returned nil")
				return
			}

			// Only inspect immutable-by-test fields from the returned
			// object immediately; no mutation is performed here.
			if leader.Address == "" {
				t.Errorf("RotateLeader returned validator with empty address")
			}
		}()
	}

	wg.Wait()

	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	if len(Validators) != validators {
		t.Fatalf("validator count changed during rotation: got %d", len(Validators))
	}

	if LeaderIndex < 0 || LeaderIndex >= len(Validators) {
		t.Fatalf(
			"LeaderIndex out of bounds: %d, validators=%d",
			LeaderIndex,
			len(Validators),
		)
	}
}

// Test concurrent jail/unjail operations on the same validator.
// The important invariant is that the validator state remains internally
// consistent and the validator list is not corrupted.
func TestConcurrentJailUnjail(t *testing.T) {
	resetValidatorConcurrencyStateForTest()
	defer resetValidatorConcurrencyStateForTest()

	const address = "0x2222222222222222222222222222222222222222"

	AddValidator(address, MinimumValidatorPower)

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()

			if i%2 == 0 {
				JailValidator(address)
			} else {
				UnjailValidator(address)
			}
		}(i)
	}

	wg.Wait()

	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	if len(Validators) != 1 {
		t.Fatalf("expected exactly 1 validator, got %d", len(Validators))
	}

	v := Validators[0]

	if v.Jailed && v.Active {
		t.Fatal("invalid validator state: validator is both jailed and active")
	}

	if !v.Jailed && !v.Active {
		t.Fatal("invalid validator state: validator is neither jailed nor active")
	}
}

// Test concurrent reads of active validator count while validators are added.
func TestConcurrentActiveValidatorCount(t *testing.T) {
	resetValidatorConcurrencyStateForTest()
	defer resetValidatorConcurrencyStateForTest()

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	for i := 0; i < workers; i++ {
		i := i

		go func() {
			defer wg.Done()

			AddValidator(
				fmt.Sprintf("0x%040x", i+1000),
				MinimumValidatorPower,
			)
		}()

		go func() {
			defer wg.Done()

			count := ActiveValidatorCount()
			if count < 0 {
				t.Errorf("active validator count became negative: %d", count)
			}
		}()
	}

	wg.Wait()

	validatorStateMu.RLock()
	defer validatorStateMu.RUnlock()

	if len(Validators) != workers {
		t.Fatalf(
			"expected %d validators, got %d",
			workers,
			len(Validators),
		)
	}

	if ActiveValidatorCount() != workers {
		t.Fatalf(
			"expected %d active validators, got %d",
			workers,
			ActiveValidatorCount(),
		)
	}
}
