package app

import (
	"errors"
	"testing"
)

func TestCommitBlockRejectsHeightGap(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       2,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "BLOCK_2",
	}

	err := CommitBlock(block)

	if !errors.Is(err, ErrBlockHeightGap) {
		t.Fatalf("expected ErrBlockHeightGap, got %v", err)
	}

	if got := len(BlockchainSnapshot()); got != 1 {
		t.Fatalf("blockchain mutated after rejected gap: got %d blocks", got)
	}
}

func TestCommitBlockRejectsPreviousHashMismatch(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       2,
		PreviousHash: "WRONG_PARENT",
		Hash:         "BLOCK_2",
	}

	err := CommitBlock(block)

	if !errors.Is(err, ErrBlockPreviousHash) {
		t.Fatalf("expected ErrBlockPreviousHash, got %v", err)
	}

	if got := len(BlockchainSnapshot()); got != 2 {
		t.Fatalf("blockchain mutated after rejected parent mismatch: got %d blocks", got)
	}
}

func TestCommitBlockAcceptsSequentialBlock(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       2,
		PreviousHash: "BLOCK_1",
		Hash:         "BLOCK_2",
	}

	if err := CommitBlock(block); err != nil {
		t.Fatalf("CommitBlock failed: %v", err)
	}

	snapshot := BlockchainSnapshot()

	if len(snapshot) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(snapshot))
	}

	if snapshot[2].Height != 2 {
		t.Fatalf("unexpected committed height: %d", snapshot[2].Height)
	}

	if snapshot[2].Hash != "BLOCK_2" {
		t.Fatalf("unexpected committed hash: %s", snapshot[2].Hash)
	}
}

func TestCommitBlockIsIdempotent(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "BLOCK_1",
	}

	err := CommitBlock(block)

	if !errors.Is(err, ErrBlockAlreadyCommitted) {
		t.Fatalf("expected ErrBlockAlreadyCommitted, got %v", err)
	}

	if got := len(BlockchainSnapshot()); got != 2 {
		t.Fatalf("idempotent commit changed chain length: got %d", got)
	}
}

func TestCommitBlockRejectsConflictingBlock(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
		{
			Height:       1,
			PreviousHash: "GENESIS_BLOCK",
			Hash:         "BLOCK_1",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "ATTACKER_BLOCK",
	}

	err := CommitBlock(block)

	if !errors.Is(err, ErrBlockConflict) {
		t.Fatalf("expected ErrBlockConflict, got %v", err)
	}

	snapshot := BlockchainSnapshot()

	if len(snapshot) != 2 {
		t.Fatalf("chain changed after conflict: got %d blocks", len(snapshot))
	}

	if snapshot[1].Hash != "BLOCK_1" {
		t.Fatalf("canonical block was replaced: %s", snapshot[1].Hash)
	}
}

func TestCommitBlockDoesNotAliasTransactions(t *testing.T) {
	blockchainMu.Lock()
	old := Blockchain
	Blockchain = []Block{
		{
			Height: 0,
			Hash:   "GENESIS_BLOCK",
		},
	}
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockchainMu.Lock()
		Blockchain = old
		blockchainMu.Unlock()
	})

	block := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Hash:         "BLOCK_1",
		Transactions: []Transaction{
			{ID: "TX_1"},
		},
	}

	if err := CommitBlock(block); err != nil {
		t.Fatalf("CommitBlock failed: %v", err)
	}

	block.Transactions[0].ID = "MUTATED"

	snapshot := BlockchainSnapshot()

	if snapshot[1].Transactions[0].ID != "TX_1" {
		t.Fatal("committed transaction state aliases caller memory")
	}
}
