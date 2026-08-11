package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupBlockchainRecoveryTest(t *testing.T) {
	t.Helper()

	blockStorageRoot = t.TempDir()

	blockchainMu.Lock()
	original := Blockchain
	Blockchain = nil
	blockchainMu.Unlock()

	t.Cleanup(func() {
		blockStorageRoot = ""

		blockchainMu.Lock()
		Blockchain = original
		blockchainMu.Unlock()
	})
}

func makeRecoveryTestBlock(
	t *testing.T,
	height int,
	previousHash string,
	tx Transaction,
) Block {
	t.Helper()

	txHash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatalf("GenerateTransactionHash failed: %v", err)
	}

	tx.Hash = txHash

	block := Block{
		Height:       height,
		PreviousHash: previousHash,
		Timestamp:    "2026-01-01T00:00:00Z",
		Transactions: []Transaction{tx},
	}

	hash, err := GenerateBlockHash(block)
	if err != nil {
		t.Fatalf("GenerateBlockHash failed: %v", err)
	}

	block.Hash = hash

	return block
}

func TestRecoverBlockchainFromDisk(t *testing.T) {
	setupBlockchainRecoveryTest(t)

	genesis := Block{
		Height: 0,
		Hash:   "GENESIS_BLOCK",
	}

	if err := SaveBlock(genesis); err != nil {
		t.Fatalf("SaveBlock genesis failed: %v", err)
	}

	tx1 := Transaction{
		ID:     "recovery-test-tx-1",
		From:   "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		To:     "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Amount: 10,
		Nonce:  1,
	}

	block1 := makeRecoveryTestBlock(
		t,
		1,
		genesis.Hash,
		tx1,
	)

	if err := SaveBlock(block1); err != nil {
		t.Fatalf("SaveBlock block1 failed: %v", err)
	}

	if err := RecoverBlockchainFromDisk(); err != nil {
		t.Fatalf("RecoverBlockchainFromDisk failed: %v", err)
	}

	snapshot := BlockchainSnapshot()

	if len(snapshot) != 2 {
		t.Fatalf("unexpected recovered block count: got %d want 2", len(snapshot))
	}

	if snapshot[0].Height != 0 {
		t.Fatalf("unexpected genesis height: %d", snapshot[0].Height)
	}

	if snapshot[1].Height != 1 {
		t.Fatalf("unexpected block height: %d", snapshot[1].Height)
	}

	if snapshot[1].PreviousHash != snapshot[0].Hash {
		t.Fatal("recovered hash chain is not linked correctly")
	}
}

func TestRecoverBlockchainRejectsGap(t *testing.T) {
	setupBlockchainRecoveryTest(t)

	genesis := Block{
		Height: 0,
		Hash:   "GENESIS_BLOCK",
	}

	if err := SaveBlock(genesis); err != nil {
		t.Fatalf("SaveBlock genesis failed: %v", err)
	}

	tx := Transaction{
		ID:     "recovery-test-tx-gap",
		From:   "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		To:     "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Amount: 10,
		Nonce:  1,
	}

	block2 := makeRecoveryTestBlock(
		t,
		2,
		"HASH_1",
		tx,
	)

	if err := SaveBlock(block2); err != nil {
		t.Fatalf("SaveBlock block2 failed: %v", err)
	}

	err := RecoverBlockchainFromDisk()

	if err == nil {
		t.Fatal("expected recovery to reject missing block height 1")
	}

	if !errors.Is(err, ErrBlockchainRecoveryGap) {
		t.Fatalf("expected ErrBlockchainRecoveryGap, got %v", err)
	}
}

func TestRecoverBlockchainRejectsInvalidBlock(t *testing.T) {
	setupBlockchainRecoveryTest(t)

	genesis := Block{
		Height: 0,
		Hash:   "GENESIS_BLOCK",
	}

	if err := SaveBlock(genesis); err != nil {
		t.Fatalf("SaveBlock genesis failed: %v", err)
	}

	tx := Transaction{
		ID:     "recovery-test-tx-invalid",
		From:   "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		To:     "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Amount: 10,
		Nonce:  1,
	}

	block1 := makeRecoveryTestBlock(
		t,
		1,
		genesis.Hash,
		tx,
	)

	block1.Hash = "CORRUPTED_HASH"

	if err := SaveBlock(block1); err != nil {
		t.Fatalf("SaveBlock corrupted block failed: %v", err)
	}

	err := RecoverBlockchainFromDisk()

	if err == nil {
		t.Fatal("expected invalid block recovery failure")
	}
}

func TestRecoverBlockchainRejectsInvalidGenesis(t *testing.T) {
	setupBlockchainRecoveryTest(t)

	dir, err := BlockStorageDir()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}

	data := []byte(`{"Height":0,"Hash":"BAD_GENESIS"}`)

	if err := os.WriteFile(
		filepath.Join(dir, "0.json"),
		data,
		0600,
	); err != nil {
		t.Fatal(err)
	}

	err = RecoverBlockchainFromDisk()

	if err == nil {
		t.Fatal("expected invalid genesis failure")
	}

	if !errors.Is(err, ErrBlockchainRecoveryGenesis) {
		t.Fatalf(
			"expected ErrBlockchainRecoveryGenesis, got %v",
			err,
		)
	}
}
