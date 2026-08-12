package app

import (
	"fmt"
	"testing"
	"time"
)

const (
	blockProducerTestFrom = "0x1111111111111111111111111111111111111111"
	blockProducerTestTo   = "0x2222222222222222222222222222222222222222"
)

func newBlockProducerTestTx(
	t *testing.T,
	id string,
	fee uint64,
	timestamp time.Time,
) Transaction {
	t.Helper()

	tx := Transaction{
		ID:        id,
		From:      blockProducerTestFrom,
		To:        blockProducerTestTo,
		Amount:    1,
		GasLimit:  DefaultGasLimit,
		GasPrice:  0,
		Fee:       fee,
		Nonce:     1,
		Timestamp: timestamp.UTC(),
	}

	hash, err := GenerateTransactionHash(tx)
	if err != nil {
		t.Fatalf("GenerateTransactionHash failed: %v", err)
	}

	tx.Hash = hash

	return tx
}

func TestProduceBlockSelectsHighestPriorityTransactions(t *testing.T) {
	oldMempool := NodeMempool
	defer func() {
		NodeMempool = oldMempool
	}()

	mempool := NewMempool()
	NodeMempool = mempool

	now := time.Now().UTC()

	low := newBlockProducerTestTx(
		t,
		"block-priority-low",
		10,
		now,
	)

	high := newBlockProducerTestTx(
		t,
		"block-priority-high",
		100,
		now.Add(time.Second),
	)

	mempool.AddTransaction(low)
	mempool.AddTransaction(high)

	block := ProduceBlock(1, "GENESIS_BLOCK")

	if len(block.Transactions) != 2 {
		t.Fatalf(
			"expected 2 transactions, got %d",
			len(block.Transactions),
		)
	}

	if block.Transactions[0].Hash != high.Hash {
		t.Fatalf(
			"highest-priority transaction was not selected first: got %s want %s",
			block.Transactions[0].Hash,
			high.Hash,
		)
	}

	if block.Transactions[1].Hash != low.Hash {
		t.Fatalf(
			"second transaction ordering incorrect: got %s want %s",
			block.Transactions[1].Hash,
			low.Hash,
		)
	}
}

func TestProduceBlockRespectsMaxTransactionsPerBlock(t *testing.T) {
	oldMempool := NodeMempool
	defer func() {
		NodeMempool = oldMempool
	}()

	mempool := NewMempool()
	NodeMempool = mempool

	for i := 0; i < MaxTransactionsPerBlock+100; i++ {
		tx := newBlockProducerTestTx(
			t,
			fmt.Sprintf("block-limit-%d", i),
			uint64(i+1),
			time.Unix(int64(i+1), 0).UTC(),
		)

		mempool.AddTransaction(tx)
	}

	block := ProduceBlock(1, "GENESIS_BLOCK")

	if len(block.Transactions) > MaxTransactionsPerBlock {
		t.Fatalf(
			"block contains too many transactions: got %d max %d",
			len(block.Transactions),
			MaxTransactionsPerBlock,
		)
	}

	if len(block.Transactions) != MaxTransactionsPerBlock {
		t.Fatalf(
			"expected block to contain %d transactions, got %d",
			MaxTransactionsPerBlock,
			len(block.Transactions),
		)
	}
}

func TestProduceBlockDoesNotMutateMempool(t *testing.T) {
	oldMempool := NodeMempool
	defer func() {
		NodeMempool = oldMempool
	}()

	mempool := NewMempool()
	NodeMempool = mempool

	tx := newBlockProducerTestTx(
		t,
		"block-persistent",
		100,
		time.Now().UTC(),
	)

	mempool.AddTransaction(tx)

	before := mempool.Count()

	block := ProduceBlock(1, "GENESIS_BLOCK")

	if block.Hash == "" {
		t.Fatal("produced block hash is empty")
	}

	if got := mempool.Count(); got != before {
		t.Fatalf(
			"ProduceBlock mutated mempool: before=%d after=%d",
			before,
			got,
		)
	}
}

func TestProduceBlockEmptyMempool(t *testing.T) {
	oldMempool := NodeMempool
	defer func() {
		NodeMempool = oldMempool
	}()

	NodeMempool = NewMempool()

	block := ProduceBlock(1, "GENESIS_BLOCK")

	if block.Hash == "" {
		t.Fatal("empty-mempool block should still have a valid hash")
	}

	if len(block.Transactions) != 0 {
		t.Fatalf(
			"expected empty transaction list, got %d",
			len(block.Transactions),
		)
	}
}

func TestProduceBlockPreservesDeterministicPriority(t *testing.T) {
	oldMempool := NodeMempool
	defer func() {
		NodeMempool = oldMempool
	}()

	mempool := NewMempool()
	NodeMempool = mempool

	timestamp := time.Unix(1000, 0).UTC()

	bbb := newBlockProducerTestTx(
		t,
		"block-bbb",
		100,
		timestamp,
	)

	aaa := newBlockProducerTestTx(
		t,
		"block-aaa",
		100,
		timestamp,
	)

	mempool.AddTransaction(bbb)
	mempool.AddTransaction(aaa)

	// The mempool comparator defines the deterministic tie-break.
	// Reproduce that expected order rather than assuming ID order.
	expectedFirst := bbb
	expectedSecond := aaa

	if aaa.Hash < bbb.Hash {
		expectedFirst = aaa
		expectedSecond = bbb
	}

	block := ProduceBlock(1, "GENESIS_BLOCK")

	if len(block.Transactions) != 2 {
		t.Fatalf(
			"expected 2 transactions, got %d",
			len(block.Transactions),
		)
	}

	if block.Transactions[0].Hash != expectedFirst.Hash {
		t.Fatalf(
			"block producer lost deterministic ordering: first=%s want %s",
			block.Transactions[0].Hash,
			expectedFirst.Hash,
		)
	}

	if block.Transactions[1].Hash != expectedSecond.Hash {
		t.Fatalf(
			"block producer lost deterministic ordering: second=%s want %s",
			block.Transactions[1].Hash,
			expectedSecond.Hash,
		)
	}
}
