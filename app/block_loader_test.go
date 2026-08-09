package app

import (
	"os"
	"path/filepath"
	"testing"
)

func setupBlockStorageTest(t *testing.T) {
	t.Helper()

	blockStorageRoot = t.TempDir()

	t.Cleanup(func() {
		blockStorageRoot = ""
	})
}

func TestGetLatestBlock(t *testing.T) {
	setupBlockStorageTest(t)

	blocks := []Block{
		{
			Height:       1,
			Hash:         "HASH_1",
			PreviousHash: "GENESIS_BLOCK",
			Timestamp:    "2026-01-01T00:00:01Z",
		},
		{
			Height:       2,
			Hash:         "HASH_2",
			PreviousHash: "HASH_1",
			Timestamp:    "2026-01-01T00:00:02Z",
		},
		{
			Height:       3,
			Hash:         "HASH_3",
			PreviousHash: "HASH_2",
			Timestamp:    "2026-01-01T00:00:03Z",
		},
	}

	for _, block := range blocks {
		if err := SaveBlock(block); err != nil {
			t.Fatalf("SaveBlock failed: %v", err)
		}
	}

	latest, err := GetLatestBlock()
	if err != nil {
		t.Fatalf("GetLatestBlock failed: %v", err)
	}

	if latest.Height != 3 {
		t.Fatalf("latest height mismatch: got %d, want 3", latest.Height)
	}

	if latest.Hash != "HASH_3" {
		t.Fatalf("latest hash mismatch: got %s, want HASH_3", latest.Hash)
	}
}

func TestLoadMissingBlock(t *testing.T) {
	setupBlockStorageTest(t)

	_, err := LoadBlock(999999)

	if err == nil {
		t.Fatal("expected error when loading missing block")
	}
}

func TestGetLatestBlockIgnoresInvalidFiles(t *testing.T) {
	setupBlockStorageTest(t)

	dir, err := BlockStorageDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	invalidFile := filepath.Join(dir, "invalid.json")

	if err := os.WriteFile(
		invalidFile,
		[]byte(`not valid json`),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	block := Block{
		Height:       5,
		Hash:         "VALID_HASH_5",
		PreviousHash: "HASH_4",
		Timestamp:    "2026-01-01T00:00:05Z",
	}

	if err := SaveBlock(block); err != nil {
		t.Fatalf("SaveBlock failed: %v", err)
	}

	latest, err := GetLatestBlock()
	if err != nil {
		t.Fatalf("GetLatestBlock failed: %v", err)
	}

	if latest.Height != 5 {
		t.Fatalf("latest height mismatch: got %d, want 5", latest.Height)
	}
}
