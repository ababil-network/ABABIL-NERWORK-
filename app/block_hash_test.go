package app

import (
	"strings"
	"testing"
)

func TestGenerateBlockHashIsDeterministic(t *testing.T) {
	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Timestamp:    "2026-08-10T00:00:00Z",
		Transactions: []Transaction{},
	}

	hash1, err := GenerateBlockHash(block)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := GenerateBlockHash(block)
	if err != nil {
		t.Fatal(err)
	}

	if hash1 == "" {
		t.Fatal("generated block hash is empty")
	}

	if !strings.EqualFold(hash1, hash2) {
		t.Fatalf("block hash is not deterministic: %s != %s", hash1, hash2)
	}
}

func TestGenerateBlockHashChangesWhenBlockChanges(t *testing.T) {
	base := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Timestamp:    "2026-08-10T00:00:00Z",
		Transactions: []Transaction{},
	}

	baseHash, err := GenerateBlockHash(base)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		block Block
	}{
		{
			name: "height",
			block: func() Block {
				b := base
				b.Height = 2
				return b
			}(),
		},
		{
			name: "previous hash",
			block: func() Block {
				b := base
				b.PreviousHash = "different-parent"
				return b
			}(),
		},
		{
			name: "timestamp",
			block: func() Block {
				b := base
				b.Timestamp = "2026-08-10T00:00:01Z"
				return b
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := GenerateBlockHash(tt.block)
			if err != nil {
				t.Fatal(err)
			}

			if strings.EqualFold(baseHash, hash) {
				t.Fatalf("block hash did not change after modifying %s", tt.name)
			}
		})
	}
}

func TestValidateBlockRejectsTamperedHash(t *testing.T) {
	previous := Block{
		Height: 0,
		Hash:   "GENESIS_BLOCK",
	}

	tx := testBlockTransaction("hash-test-tx", "")

	txHash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = txHash

	block := Block{
		Height:       1,
		PreviousHash: previous.Hash,
		Timestamp:    "2026-08-10T00:00:00Z",
		Transactions: []Transaction{tx},
	}

	block.Hash, err = GenerateBlockHash(block)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateBlock(block, previous); err != nil {
		t.Fatalf("valid block rejected: %v", err)
	}

	block.Hash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if err := ValidateBlock(block, previous); err == nil {
		t.Fatal("tampered block hash was accepted")
	}
}
