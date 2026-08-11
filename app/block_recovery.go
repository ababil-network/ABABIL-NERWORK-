package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

var (
	ErrBlockchainRecoveryEmpty   = errors.New("no persisted blockchain found")
	ErrBlockchainRecoveryGap     = errors.New("blockchain height gap detected")
	ErrBlockchainRecoveryGenesis = errors.New("invalid persisted genesis block")
)

func RecoverBlockchainFromDisk() error {
	dir, err := BlockStorageDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrBlockchainRecoveryEmpty
		}
		return err
	}

	heights := make([]int, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if filepath.Ext(name) != ".json" {
			continue
		}

		base := name[:len(name)-len(filepath.Ext(name))]

		height, err := strconv.Atoi(base)
		if err != nil || height < 0 {
			continue
		}

		heights = append(heights, height)
	}

	if len(heights) == 0 {
		return ErrBlockchainRecoveryEmpty
	}

	sort.Ints(heights)

	// Recovery must always begin from the genesis block.
	if heights[0] != 0 {
		return fmt.Errorf("%w: genesis block is missing", ErrBlockchainRecoveryGenesis)
	}

	// Reject duplicate logical heights. Normally impossible with filenames,
	// but keep this invariant explicit.
	for i := 1; i < len(heights); i++ {
		if heights[i] == heights[i-1] {
			return fmt.Errorf("%w: duplicate height %d", ErrBlockchainRecoveryGap, heights[i])
		}
	}

	recovered := make([]Block, 0, len(heights))

	// Load and validate genesis separately because ValidateBlock requires
	// normal blocks to contain at least one transaction.
	genesis, err := LoadBlock(0)
	if err != nil {
		return fmt.Errorf("failed to load genesis block: %w", err)
	}

	if genesis.Height != 0 ||
		genesis.Hash != "GENESIS_BLOCK" ||
		genesis.PreviousHash != "" {
		return fmt.Errorf(
			"%w: height=%d hash=%s",
			ErrBlockchainRecoveryGenesis,
			genesis.Height,
			genesis.Hash,
		)
	}

	recovered = append(recovered, genesis)

	previous := genesis

	for expectedHeight := 1; expectedHeight <= heights[len(heights)-1]; expectedHeight++ {
		block, err := LoadBlock(expectedHeight)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(
					"%w: missing height %d",
					ErrBlockchainRecoveryGap,
					expectedHeight,
				)
			}

			return fmt.Errorf(
				"failed to load block %d: %w",
				expectedHeight,
				err,
			)
		}

		if err := ValidateBlock(block, previous); err != nil {
			return fmt.Errorf(
				"invalid persisted block %d: %w",
				expectedHeight,
				err,
			)
		}

		recovered = append(recovered, block)
		previous = block
	}

	// Only mutate the live blockchain after the entire persisted chain
	// has been successfully validated.
	blockchainMu.Lock()
	Blockchain = recovered
	blockchainMu.Unlock()

	LogInfo("=================================")
	LogInfo("Blockchain Recovered")
	LogInfo(fmt.Sprintf("Blocks : %d", len(recovered)))
	LogInfo(fmt.Sprintf("Height : %d", previous.Height))
	LogInfo("Hash   : " + previous.Hash)
	LogInfo("=================================")

	return nil
}
