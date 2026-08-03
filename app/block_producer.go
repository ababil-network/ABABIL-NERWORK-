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

func ProduceBlock(height int, previousHash string, mempool *Mempool) Block {

	block := Block{
		Height:       height,
		Hash:         GenerateBlockHash(),
		PreviousHash: previousHash,
		Timestamp:    time.Now().Format(time.RFC3339),
		Transactions: mempool.Transactions,
	}

	// Clear mempool after block creation
	mempool.Transactions = nil
leader := GetLeader()

if leader != nil {
	LogInfo("=================================")
	LogInfo("Block Produced")
	LogInfo("Leader : " + leader.Address)
	LogInfo("=================================")
}
	return block
}
