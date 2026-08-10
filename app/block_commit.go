package app

import (
	"errors"
	"sync"
)

var (
	ErrBlockAlreadyCommitted = errors.New("block already committed")
	ErrBlockConflict         = errors.New("conflicting block at committed height")
)

var (
	blockchainMu sync.RWMutex
	Blockchain   []Block
)

// CommitBlock records a block in the in-memory chain.
//
// The block must already have passed structural validation and, for network
// blocks, must already have been persisted successfully.
func CommitBlock(block Block) error {
	blockchainMu.Lock()
	defer blockchainMu.Unlock()

	for _, existing := range Blockchain {
		if existing.Height != block.Height {
			continue
		}

		if existing.Hash == block.Hash {
			return ErrBlockAlreadyCommitted
		}

		return ErrBlockConflict
	}

	// Copy the transaction slice so callers cannot mutate committed state
	// through their original block value.
	if block.Transactions != nil {
		block.Transactions = append([]Transaction(nil), block.Transactions...)
	}

	Blockchain = append(Blockchain, block)

	LogInfo("=================================")
	LogInfo("Block Committed")
	LogInfo("Block Height Recorded")
	LogInfo("Hash   : " + block.Hash)
	LogInfo("=================================")

	return nil
}

// BlockchainSnapshot returns a safe copy of the in-memory blockchain.
func BlockchainSnapshot() []Block {
	blockchainMu.RLock()
	defer blockchainMu.RUnlock()

	snapshot := make([]Block, len(Blockchain))

	for i, block := range Blockchain {
		snapshot[i] = block

		if block.Transactions != nil {
			snapshot[i].Transactions = append(
				[]Transaction(nil),
				block.Transactions...,
			)
		}
	}

	return snapshot
}
