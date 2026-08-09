package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveBlock(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	block := Block{
		Height:       1,
		Hash:         "TEST_BLOCK_HASH",
		PreviousHash: "GENESIS_BLOCK",
		Timestamp:    "2026-01-01T00:00:00Z",
		Transactions: []Transaction{},
	}

	if err := SaveBlock(block); err != nil {
		t.Fatalf("SaveBlock failed: %v", err)
	}

	file := filepath.Join(
		home,
		".ababil",
		"data",
		"blocks",
		"1.json",
	)

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("saved block file not found: %v", err)
	}

	var saved Block

	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("invalid saved block JSON: %v", err)
	}

	if saved.Height != block.Height {
		t.Fatalf("height mismatch: got %d, want %d", saved.Height, block.Height)
	}

	if saved.Hash != block.Hash {
		t.Fatalf("hash mismatch: got %s, want %s", saved.Hash, block.Hash)
	}

	if saved.PreviousHash != block.PreviousHash {
		t.Fatalf(
			"previous hash mismatch: got %s, want %s",
			saved.PreviousHash,
			block.PreviousHash,
		)
	}
}
func TestLoadBlock(t *testing.T) {
	block := Block{
		Height:       2,
		Hash:         "LOAD_TEST_HASH",
		PreviousHash: "TEST_BLOCK_HASH",
		Timestamp:    "2026-01-01T00:00:00Z",
		Transactions: []Transaction{},
	}

	if err := SaveBlock(block); err != nil {
		t.Fatalf("SaveBlock failed: %v", err)
	}

	loaded, err := LoadBlock(block.Height)
	if err != nil {
		t.Fatalf("LoadBlock failed: %v", err)
	}

	if loaded.Height != block.Height {
		t.Fatalf("height mismatch: got %d, want %d", loaded.Height, block.Height)
	}

	if loaded.Hash != block.Hash {
		t.Fatalf("hash mismatch: got %s, want %s", loaded.Hash, block.Hash)
	}

	if loaded.PreviousHash != block.PreviousHash {
		t.Fatalf(
			"previous hash mismatch: got %s, want %s",
			loaded.PreviousHash,
			block.PreviousHash,
		)
	}
}
func TestLoadCorruptedBlock(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(
		home,
		".ababil",
		"data",
		"blocks",
	)

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	file := filepath.Join(dir, "999.json")

	if err := os.WriteFile(
		file,
		[]byte(`{"height":`),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(file)

	_, err = LoadBlock(999)

	if err == nil {
		t.Fatal("corrupted block was loaded successfully")
	}
}
