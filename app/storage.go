package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	ErrBlockStorageConflict = errors.New("conflicting block already exists at height")
)

func SaveBlock(block Block) error {
	dir, err := BlockStorageDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file := filepath.Join(dir, fmt.Sprintf("%d.json", block.Height))

	// Never silently overwrite an existing block.
	if existingData, err := os.ReadFile(file); err == nil {
		var existing Block

		if err := json.Unmarshal(existingData, &existing); err != nil {
			return fmt.Errorf("existing block file is invalid: %w", err)
		}

		if existing.Hash == block.Hash {
			// Idempotent save.
			return nil
		}

		return fmt.Errorf(
			"%w: height=%d existing=%s incoming=%s",
			ErrBlockStorageConflict,
			block.Height,
			existing.Hash,
			block.Hash,
		)
	} else if !os.IsNotExist(err) {
		return err
	}

	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return err
	}

	temp := file + ".tmp"

	f, err := os.OpenFile(
		temp,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return err
	}

	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(temp)
	}

	if _, err := f.Write(data); err != nil {
		cleanup()
		return err
	}

	if err := f.Sync(); err != nil {
		cleanup()
		return err
	}

	if err := f.Close(); err != nil {
		_ = os.Remove(temp)
		return err
	}

	if err := os.Rename(temp, file); err != nil {
		_ = os.Remove(temp)
		return err
	}

	return nil
}
