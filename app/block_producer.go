package app

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func GenerateBlockHash() string {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func ProduceBlock(height int, previousHash string) Block {
	txs := NodeMempool.Transactions

	if len(txs) > MaxTransactionsPerBlock {
		txs = txs[:MaxTransactionsPerBlock]
	}
	block := Block{
		Height:       height,
		Hash:         GenerateBlockHash(),
		PreviousHash: previousHash,
		Timestamp:    time.Now().Format(time.RFC3339),
		Transactions: txs,
	}

	NodeMempool.RemoveProcessedTransactions(block.Transactions)

	leader := GetLeader()

	if leader != nil {
		LogInfo("=================================")
		LogInfo("Block Produced")
		LogInfo("Leader : " + leader.Address)
		LogInfo("=================================")
	}
	return block
}
