package app

import (
	"errors"
	"sync"
)

var (
	ErrBlockAlreadyCommitted = errors.New("block already committed")
	ErrBlockConflict         = errors.New("conflicting block at committed height")
	ErrBlockHeightGap        = errors.New("block height gap")
	ErrBlockPreviousHash     = errors.New("block previous hash mismatch")
	ErrInvalidGenesisCommit  = errors.New("invalid genesis commit")
)

var (
	blockchainMu sync.RWMutex
	Blockchain   []Block
)

// CommitBlock records a block in the in-memory chain.
//
// CommitBlock is the final in-memory commit gate. It enforces:
//
//   - genesis may only be committed as height 0
//   - normal blocks must extend the current tip by exactly one height
//   - PreviousHash must match the current tip hash
//   - identical blocks are idempotent
//   - conflicting blocks at an existing height are rejected
//   - committed transaction slices are copied defensively
//
// Structural validation and durable persistence are expected to happen
// before calling CommitBlock for normal network blocks.
func CommitBlock(block Block) error {
	blockchainMu.Lock()
	defer blockchainMu.Unlock()

	// Genesis is a special block and must never be treated as a normal block.
	if block.Height == 0 {
		if block.Hash != "GENESIS_BLOCK" ||
			block.PreviousHash != "" ||
			len(block.Transactions) != 0 {
			return ErrInvalidGenesisCommit
		}

		if len(Blockchain) == 0 {
			Blockchain = append(Blockchain, cloneBlock(block))

			LogInfo("=================================")
			LogInfo("Genesis Block Committed")
			LogInfo("Hash   : " + block.Hash)
			LogInfo("=================================")

			return nil
		}

		existing := Blockchain[0]

		if existing.Hash == block.Hash {
			return ErrBlockAlreadyCommitted
		}

		return ErrBlockConflict
	}

	// A normal block cannot be committed before genesis.
	if len(Blockchain) == 0 {
		return ErrBlockHeightGap
	}

	latest := Blockchain[len(Blockchain)-1]

	// A block already present at the current/previous height is handled
	// idempotently or rejected as a conflict.
	if block.Height <= latest.Height {
		for _, existing := range Blockchain {
			if existing.Height != block.Height {
				continue
			}

			if existing.Hash == block.Hash {
				return ErrBlockAlreadyCommitted
			}

			return ErrBlockConflict
		}

		return ErrBlockHeightGap
	}

	// Exactly one height must be added at a time.
	if block.Height != latest.Height+1 {
		return ErrBlockHeightGap
	}

	// The new block must extend the current canonical tip.
	if block.PreviousHash != latest.Hash {
		return ErrBlockPreviousHash
	}

	Blockchain = append(Blockchain, cloneBlock(block))

	LogInfo("=================================")
	LogInfo("Block Committed")
	LogInfo("Block Height Recorded")
	LogInfo("Hash   : " + block.Hash)
	LogInfo("=================================")

	return nil
}

// cloneBlock returns a defensive copy of a block.
func cloneBlock(block Block) Block {
	cloned := block

	if block.Transactions != nil {
		cloned.Transactions = append(
			[]Transaction(nil),
			block.Transactions...,
		)
	}

	return cloned
}

// BlockchainSnapshot returns a safe copy of the in-memory blockchain.
func BlockchainSnapshot() []Block {
	blockchainMu.RLock()
	defer blockchainMu.RUnlock()

	snapshot := make([]Block, len(Blockchain))

	for i, block := range Blockchain {
		snapshot[i] = cloneBlock(block)
	}

	return snapshot
}
