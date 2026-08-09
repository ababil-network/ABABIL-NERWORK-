package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func SaveBlock(block Block) error {
	dir, err := BlockStorageDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return err
	}

	file := filepath.Join(dir, fmt.Sprintf("%d.json", block.Height))
	temp := file + ".tmp"

	f, err := os.OpenFile(
		temp,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
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
