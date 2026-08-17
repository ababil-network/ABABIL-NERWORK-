package app

import "time"

func ProduceBlock(height int, previousHash string) Block {
	txs := NodeMempool.Snapshot()

	if len(txs) > MaxTransactionsPerBlock {
		txs = txs[:MaxTransactionsPerBlock]
	}

	leader := GetLeader()
	if leader == nil {
		LogError("failed to produce block: no eligible proposer")
		return Block{}
	}

	block := Block{
		Height:       height,
		PreviousHash: previousHash,
		Proposer:     leader.Address,
		Timestamp:    time.Now().UTC().Format(time.RFC3339Nano),
		Transactions: txs,
	}

	hash, err := GenerateBlockHash(block)
	if err != nil {
		LogError("failed to generate block hash: " + err.Error())
		return Block{}
	}

	block.Hash = hash

	LogInfo("=================================")
	LogInfo("Block Produced")
	LogInfo("Proposer : " + block.Proposer)
	LogInfo("=================================")

	return block
}
