package app

import "time"

func ProduceBlock(height int, previousHash string) Block {
	txs := NodeMempool.Snapshot()

	if len(txs) > MaxTransactionsPerBlock {
		txs = txs[:MaxTransactionsPerBlock]
	}

	block := Block{
		Height:       height,
		PreviousHash: previousHash,
		Timestamp:    time.Now().UTC().Format(time.RFC3339Nano),
		Transactions: txs,
	}

	hash, err := GenerateBlockHash(block)
	if err != nil {
		LogError("failed to generate block hash: " + err.Error())
		return Block{}
	}

	block.Hash = hash

	leader := GetLeader()

	if leader != nil {
		LogInfo("=================================")
		LogInfo("Block Produced")
		LogInfo("Leader : " + leader.Address)
		LogInfo("=================================")
	}

	return block
}
