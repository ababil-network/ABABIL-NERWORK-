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
