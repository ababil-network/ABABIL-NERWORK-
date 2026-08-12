package app

import (
	"testing"
	"time"
)

func TestValidateBlockRequest(t *testing.T) {
	tests := []struct {
		name    string
		request BlockRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: BlockRequest{
				FromHeight: 1,
				ToHeight:   10,
			},
			valid: true,
		},
		{
			name: "invalid from height",
			request: BlockRequest{
				FromHeight: 0,
				ToHeight:   10,
			},
			valid: false,
		},
		{
			name: "invalid height range",
			request: BlockRequest{
				FromHeight: 10,
				ToHeight:   5,
			},
			valid: false,
		},
		{
			name: "too many blocks",
			request: BlockRequest{
				FromHeight: 1,
				ToHeight:   1000001,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlockRequest(tt.request)

			if tt.valid && err != nil {
				t.Fatalf("expected valid request, got error: %v", err)
			}

			if !tt.valid && err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestHandleBlockRequest(t *testing.T) {
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
			t.Fatalf("failed to save block %d: %v", block.Height, err)
		}
	}

	response, err := HandleBlockRequest(BlockRequest{
		FromHeight: 1,
		ToHeight:   3,
	})
	if err != nil {
		t.Fatalf("HandleBlockRequest failed: %v", err)
	}

	if len(response.Blocks) != 3 {
		t.Fatalf("block count mismatch: got %d, want 3", len(response.Blocks))
	}

	if response.Blocks[0].Height != 1 {
		t.Fatalf("first block height mismatch: got %d, want 1", response.Blocks[0].Height)
	}

	if response.Blocks[2].Height != 3 {
		t.Fatalf("last block height mismatch: got %d, want 3", response.Blocks[2].Height)
	}
}

func TestCommitSyncedBlocks(t *testing.T) {
	setupBlockStorageTest(t)

	Blockchain = nil

	t.Cleanup(func() {
		Blockchain = nil
	})

	genesis := Block{
		Height:    0,
		Hash:      "GENESIS_BLOCK",
		Timestamp: "2026-01-01T00:00:00Z",
	}

	if err := SaveBlock(genesis); err != nil {
		t.Fatalf("failed to save genesis block: %v", err)
	}

	tx1 := Transaction{
		ID:        "SYNC_TX_1",
		From:      "0x1111111111111111111111111111111111111111",
		To:        "0x2222222222222222222222222222222222222222",
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       DefaultGasLimit * DefaultGasPrice,
		Nonce:     1,
		Timestamp: time.Unix(1700000000, 0).UTC(),
	}

	tx1Hash, err := GenerateTransactionHash(tx1)
	if err != nil {
		t.Fatalf("failed to hash tx1: %v", err)
	}
	tx1.Hash = tx1Hash

	block1 := Block{
		Height:       1,
		PreviousHash: "GENESIS_BLOCK",
		Timestamp:    "2026-01-01T00:00:01Z",
		Transactions: []Transaction{tx1},
	}

	block1Hash, err := GenerateBlockHash(block1)
	if err != nil {
		t.Fatalf("failed to hash block1: %v", err)
	}
	block1.Hash = block1Hash

	tx2 := Transaction{
		ID:        "SYNC_TX_2",
		From:      "0x3333333333333333333333333333333333333333",
		To:        "0x4444444444444444444444444444444444444444",
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       DefaultGasLimit * DefaultGasPrice,
		Nonce:     1,
		Timestamp: time.Unix(1700000001, 0).UTC(),
	}

	tx2Hash, err := GenerateTransactionHash(tx2)
	if err != nil {
		t.Fatalf("failed to hash tx2: %v", err)
	}
	tx2.Hash = tx2Hash

	block2 := Block{
		Height:       2,
		PreviousHash: block1.Hash,
		Timestamp:    "2026-01-01T00:00:02Z",
		Transactions: []Transaction{tx2},
	}

	block2Hash, err := GenerateBlockHash(block2)
	if err != nil {
		t.Fatalf("failed to hash block2: %v", err)
	}
	block2.Hash = block2Hash

	blocks := []Block{block1, block2}

	if err := CommitSyncedBlocks(blocks); err != nil {
		t.Fatalf("CommitSyncedBlocks failed: %v", err)
	}

	if len(Blockchain) != 3 {
		t.Fatalf("committed block count mismatch: got %d, want 3", len(Blockchain))
	}

	latest, err := GetLatestBlock()
	if err != nil {
		t.Fatalf("GetLatestBlock failed: %v", err)
	}

	if latest.Height != 2 {
		t.Fatalf("latest block height mismatch: got %d, want 2", latest.Height)
	}

	if latest.Hash != block2.Hash {
		t.Fatalf("latest block hash mismatch: got %s, want %s", latest.Hash, block2.Hash)
	}
}
