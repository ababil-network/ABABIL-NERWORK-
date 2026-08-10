package app

import (
	"testing"
	"time"
)

func testBlockTransaction(id, hash string) Transaction {
	return Transaction{
		ID:        id,
		From:      "0x1111111111111111111111111111111111111111",
		To:        "0x2222222222222222222222222222222222222222",
		Amount:    100,
		GasLimit:  DefaultGasLimit,
		GasPrice:  DefaultGasPrice,
		Fee:       DefaultGasLimit * DefaultGasPrice,
		Nonce:     1,
		Hash:      hash,
		Timestamp: testTransactionTimestamp(),
	}
}

func testTransactionTimestamp() time.Time {
	return time.Unix(1700000000, 0).UTC()
}

func testBlockTimestamp() string {
	return "2026-08-10T00:00:00Z"
}

func makeValidTestBlock(t *testing.T, previous Block) Block {
	t.Helper()

	tx := testBlockTransaction("test-tx-1", "")
	txHash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = txHash

	block := Block{
		Height:       previous.Height + 1,
		PreviousHash: previous.Hash,
		Timestamp:    testBlockTimestamp(),
		Transactions: []Transaction{tx},
	}

	blockHash, err := GenerateBlockHash(block)
	if err != nil {
		t.Fatal(err)
	}

	block.Hash = blockHash

	return block
}

func TestValidateBlock(t *testing.T) {
	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	validBlock := makeValidTestBlock(t, previous)

	if err := ValidateBlock(validBlock, previous); err != nil {
		t.Fatalf("valid block rejected: %v", err)
	}
}

func TestInvalidBlockHeight(t *testing.T) {
	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := makeValidTestBlock(t, previous)
	block.Height = 5

	if err := ValidateBlock(block, previous); err == nil {
		t.Fatal("invalid height accepted")
	}
}

func TestInvalidPreviousHash(t *testing.T) {
	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := makeValidTestBlock(t, previous)
	block.PreviousHash = "wrong_hash"

	if err := ValidateBlock(block, previous); err == nil {
		t.Fatal("invalid previous hash accepted")
	}
}

func TestEmptyBlock(t *testing.T) {
	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	block := Block{
		Height:       2,
		PreviousHash: "previous_hash",
		Timestamp:    testBlockTimestamp(),
		Hash:         "block_hash",
		Transactions: []Transaction{},
	}

	if err := ValidateBlock(block, previous); err == nil {
		t.Fatal("empty block accepted")
	}
}

func TestDuplicateTransaction(t *testing.T) {
	previous := Block{
		Height: 1,
		Hash:   "previous_hash",
	}

	tx := testBlockTransaction("test-tx-1", "")
	txHash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatal(err)
	}

	tx.Hash = txHash

	block := Block{
		Height:       2,
		PreviousHash: "previous_hash",
		Timestamp:    testBlockTimestamp(),
		Transactions: []Transaction{tx, tx},
	}

	block.Hash, err = GenerateBlockHash(block)
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidateBlock(block, previous); err == nil {
		t.Fatal("duplicate transaction accepted")
	}
}
