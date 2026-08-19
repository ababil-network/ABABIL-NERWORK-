package app

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrBlockProductionNoTip    = errors.New("block production requires a canonical tip")
	ErrBlockProductionNoLeader = errors.New("block production requires an eligible leader")
)

var (
	blockProductionMu      sync.Mutex
	blockProductionRunning bool
)

const BlockProductionInterval = 1 * time.Second

// StartBlockProductionWorker starts the local block-production scheduler.
//
// The worker is deliberately conservative:
//   - it never imposes a fixed transaction-count cap;
//   - block validity remains enforced by ValidateBlock;
//   - CommitBlock remains the canonical in-memory commit gate;
//   - SaveBlock remains the durable persistence gate.
//
// Capacity can later evolve to gas/byte/execution-budget based limits
// without changing the scheduler architecture.
func StartBlockProductionWorker() {
	blockProductionMu.Lock()
	if blockProductionRunning {
		blockProductionMu.Unlock()
		return
	}

	blockProductionRunning = true
	blockProductionMu.Unlock()

	go func() {
		ticker := time.NewTicker(BlockProductionInterval)
		defer ticker.Stop()

		for range ticker.C {
			if err := ProduceAndCommitNextBlock(); err != nil {
				LogError("block production cycle failed: " + err.Error())
			}
		}
	}()
}

// ProduceAndCommitNextBlock produces, validates, persists and commits
// exactly one next canonical block.
func ProduceAndCommitNextBlock() error {
	tip, err := latestCanonicalTip()
	if err != nil {
		return err
	}

	leader := GetLeader()
	if leader == nil {
		return ErrBlockProductionNoLeader
	}

	block := ProduceBlock(tip.Height+1, tip.Hash)
	if block.Hash == "" {
		return errors.New("block production returned empty block")
	}

	if block.Proposer != leader.Address {
		return fmt.Errorf(
			"block proposer mismatch: expected=%s actual=%s",
			leader.Address,
			block.Proposer,
		)
	}

	if err := ValidateBlock(block, tip); err != nil {
		return fmt.Errorf("produced block validation failed: %w", err)
	}

	if err := SaveBlock(block); err != nil {
		return fmt.Errorf("failed to persist produced block: %w", err)
	}

	if err := CommitBlock(block); err != nil {
		return fmt.Errorf("failed to commit produced block: %w", err)
	}

	if NodeMempool != nil {
		NodeMempool.RemoveProcessedTransactions(block.Transactions)
	}

	RotateLeader()

	return nil
}

func latestCanonicalTip() (Block, error) {
	snapshot := BlockchainSnapshot()
	if len(snapshot) > 0 {
		return snapshot[len(snapshot)-1], nil
	}

	tip, err := GetLatestBlock()
	if err != nil {
		return Block{}, fmt.Errorf("%w: %v", ErrBlockProductionNoTip, err)
	}

	return tip, nil
}
