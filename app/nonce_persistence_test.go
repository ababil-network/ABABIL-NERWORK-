package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersistentNonceStateRoundTrip(t *testing.T) {
	original := NodeNonce.nonces
	defer func() {
		NodeNonce.mu.Lock()
		NodeNonce.nonces = original
		NodeNonce.mu.Unlock()
	}()

	NodeNonce.mu.Lock()
	NodeNonce.nonces = map[string]uint64{
		"0x1111111111111111111111111111111111111111": 7,
		"0x2222222222222222222222222222222222222222": 19,
	}
	NodeNonce.mu.Unlock()

	state := BuildPersistentNonceState()

	if err := ValidatePersistentNonceState(state); err != nil {
		t.Fatalf("nonce state validation failed: %v", err)
	}

	if len(state.Nonces) != 2 {
		t.Fatalf("expected 2 nonce entries, got %d", len(state.Nonces))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(home, ".ababil", "data", "state", "nonces.json")
	_ = os.Remove(path)

	if err := SavePersistentNonceState(state); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadPersistentNonceState()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Nonces) != 2 {
		t.Fatalf("expected 2 loaded nonce entries, got %d", len(loaded.Nonces))
	}

	if err := ApplyPersistentNonceState(loaded); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	if got := NodeNonce.Get("0x1111111111111111111111111111111111111111"); got != 7 {
		t.Fatalf("expected nonce 7, got %d", got)
	}

	if got := NodeNonce.Get("0x2222222222222222222222222222222222222222"); got != 19 {
		t.Fatalf("expected nonce 19, got %d", got)
	}
}

func TestPersistentNonceStateRejectsDuplicateAddresses(t *testing.T) {
	state := PersistentNonceState{
		Version: CurrentNonceStateVersion,
		Nonces: []PersistentNonce{
			{Address: "0x1111111111111111111111111111111111111111", Nonce: 1},
			{Address: "0x1111111111111111111111111111111111111111", Nonce: 2},
		},
	}

	if err := ValidatePersistentNonceState(state); err != ErrNonceStateCorrupted {
		t.Fatalf("expected ErrNonceStateCorrupted, got %v", err)
	}
}

func TestPersistentNonceStateRejectsEmptyAddress(t *testing.T) {
	state := PersistentNonceState{
		Version: CurrentNonceStateVersion,
		Nonces: []PersistentNonce{
			{Address: "", Nonce: 1},
		},
	}

	if err := ValidatePersistentNonceState(state); err != ErrNonceStateCorrupted {
		t.Fatalf("expected ErrNonceStateCorrupted, got %v", err)
	}
}

func TestPersistentNonceStateRejectsWrongVersion(t *testing.T) {
	state := PersistentNonceState{
		Version: 999,
	}

	if err := ValidatePersistentNonceState(state); err != ErrNonceStateCorrupted {
		t.Fatalf("expected ErrNonceStateCorrupted, got %v", err)
	}
}
