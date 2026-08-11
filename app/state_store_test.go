package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersistentStateRoundTrip(t *testing.T) {
	original := WalletBalances

	defer func() {
		walletBalanceMu.Lock()
		WalletBalances = original
		walletBalanceIndex = nil
		walletBalanceMu.Unlock()
	}()

	WalletBalances = []WalletBalance{
		{
			Address: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Balance: 200,
		},
		{
			Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Balance: 100,
		},
	}

	walletBalanceIndex = nil

	state := BuildPersistentState(42)

	if err := SavePersistentState(state); err != nil {
		t.Fatalf("SavePersistentState failed: %v", err)
	}

	loaded, err := LoadPersistentState()
	if err != nil {
		t.Fatalf("LoadPersistentState failed: %v", err)
	}

	if loaded.Version != CurrentStateVersion {
		t.Fatalf(
			"unexpected state version: got %d want %d",
			loaded.Version,
			CurrentStateVersion,
		)
	}

	if loaded.Height != 42 {
		t.Fatalf(
			"unexpected state height: got %d want 42",
			loaded.Height,
		)
	}

	if len(loaded.Wallets) != 2 {
		t.Fatalf(
			"unexpected wallet count: got %d want 2",
			len(loaded.Wallets),
		)
	}
}

func TestPersistentStateChecksumDetectsCorruption(t *testing.T) {
	state := PersistentState{
		Version: CurrentStateVersion,
		Height:  10,
		Wallets: []PersistentWalletState{
			{
				Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Balance: 100,
			},
		},
	}

	if err := SavePersistentState(state); err != nil {
		t.Fatalf("SavePersistentState failed: %v", err)
	}

	path, err := stateFilePath()
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("state file is empty")
	}

	data[len(data)-2] ^= 0x01

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadPersistentState(); err != ErrStateChecksum &&
		err != ErrStateCorrupted {
		t.Fatalf(
			"expected corruption detection, got %v",
			err,
		)
	}
}

func TestPersistentStateRejectsDuplicateWallets(t *testing.T) {
	state := PersistentState{
		Version: CurrentStateVersion,
		Height:  1,
		Wallets: []PersistentWalletState{
			{
				Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Balance: 100,
			},
			{
				Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Balance: 200,
			},
		},
	}

	if err := ValidatePersistentState(state); err != ErrStateCorrupted {
		t.Fatalf(
			"expected ErrStateCorrupted, got %v",
			err,
		)
	}
}

func TestPersistentStateApply(t *testing.T) {
	original := WalletBalances

	defer func() {
		walletBalanceMu.Lock()
		WalletBalances = original
		walletBalanceIndex = nil
		walletBalanceMu.Unlock()
	}()

	state := PersistentState{
		Version: CurrentStateVersion,
		Height:  5,
		Wallets: []PersistentWalletState{
			{
				Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Balance: 1234,
			},
		},
	}

	if err := ApplyPersistentState(state); err != nil {
		t.Fatalf("ApplyPersistentState failed: %v", err)
	}

	if got := GetBalance(
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	); got != 1234 {
		t.Fatalf(
			"unexpected balance after state apply: got %d want 1234",
			got,
		)
	}
}

func TestPersistentStateUsesStableWalletOrdering(t *testing.T) {
	original := WalletBalances

	defer func() {
		walletBalanceMu.Lock()
		WalletBalances = original
		walletBalanceIndex = nil
		walletBalanceMu.Unlock()
	}()

	WalletBalances = []WalletBalance{
		{
			Address: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Balance: 2,
		},
		{
			Address: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Balance: 1,
		},
	}

	walletBalanceIndex = nil

	state := BuildPersistentState(1)

	if len(state.Wallets) != 2 {
		t.Fatalf("unexpected wallet count: %d", len(state.Wallets))
	}

	if state.Wallets[0].Address !=
		"0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("wallet state is not deterministically ordered")
	}
}

func TestPersistentStatePathIsInsideABABILData(t *testing.T) {
	path, err := stateFilePath()
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "state.json" {
		t.Fatalf("unexpected state file: %s", path)
	}
}
