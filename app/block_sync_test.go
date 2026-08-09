package app

import "testing"

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
				ToHeight:   MaxBlocksPerSyncRequest + 1,
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
			Hash:         "SYNC_HASH_1",
			PreviousHash: "GENESIS_BLOCK",
			Timestamp:    "2026-01-01T00:00:01Z",
			Transactions: []Transaction{
				{
					ID:   "SYNC_TX_1",
					Hash: "SYNC_TX_HASH_1",
				},
			},
		},
		{
			Height:       2,
			Hash:         "SYNC_HASH_2",
			PreviousHash: "SYNC_HASH_1",
			Timestamp:    "2026-01-01T00:00:02Z",
			Transactions: []Transaction{
				{
					ID:   "SYNC_TX_2",
					Hash: "SYNC_TX_HASH_2",
				},
			},
		},
		{
			Height:       3,
			Hash:         "SYNC_HASH_3",
			PreviousHash: "SYNC_HASH_2",
			Timestamp:    "2026-01-01T00:00:03Z",
			Transactions: []Transaction{
				{
					ID:   "SYNC_TX_3",
					Hash: "SYNC_TX_HASH_3",
				},
			},
		},
	}

	for _, block := range blocks {
		if err := SaveBlock(block); err != nil {
			t.Fatalf("SaveBlock failed: %v", err)
		}
	}

	req := BlockRequest{
		FromHeight: 1,
		ToHeight:   3,
	}

	response, err := HandleBlockRequest(req)
	if err != nil {
		t.Fatalf("HandleBlockRequest failed: %v", err)
	}

	if len(response.Blocks) != 3 {
		t.Fatalf(
			"block count mismatch: got %d, want 3",
			len(response.Blocks),
		)
	}

	if response.Blocks[0].Height != 1 {
		t.Fatalf(
			"first block height mismatch: got %d, want 1",
			response.Blocks[0].Height,
		)
	}

	if response.Blocks[2].Height != 3 {
		t.Fatalf(
			"last block height mismatch: got %d, want 3",
			response.Blocks[2].Height,
		)
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

	blocks := []Block{
		{
			Height:       1,
			Hash:         "SYNC_HASH_1",
			PreviousHash: "GENESIS_BLOCK",
			Timestamp:    "2026-01-01T00:00:01Z",
			Transactions: []Transaction{
				{
					ID:   "SYNC_TX_1",
					Hash: "SYNC_TX_HASH_1",
				},
			},
		},
		{
			Height:       2,
			Hash:         "SYNC_HASH_2",
			PreviousHash: "SYNC_HASH_1",
			Timestamp:    "2026-01-01T00:00:02Z",
			Transactions: []Transaction{
				{
					ID:   "SYNC_TX_2",
					Hash: "SYNC_TX_HASH_2",
				},
			},
		},
	}

	if err := CommitSyncedBlocks(blocks); err != nil {
		t.Fatalf("CommitSyncedBlocks failed: %v", err)
	}

	if len(Blockchain) != 2 {
		t.Fatalf(
			"committed block count mismatch: got %d, want 2",
			len(Blockchain),
		)
	}

	latest, err := GetLatestBlock()
	if err != nil {
		t.Fatalf("GetLatestBlock failed: %v", err)
	}

	if latest.Height != 2 {
		t.Fatalf(
			"latest block height mismatch: got %d, want 2",
			latest.Height,
		)
	}

	if latest.Hash != "SYNC_HASH_2" {
		t.Fatalf(
			"latest block hash mismatch: got %s, want SYNC_HASH_2",
			latest.Hash,
		)
	}
}
