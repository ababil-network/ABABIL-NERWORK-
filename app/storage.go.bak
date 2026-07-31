package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveBlock(block Block) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".ababil", "data", "blocks")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(block, "", "  ")
	if err != nil {
		return err
	}

	file := filepath.Join(dir, "genesis.json")

	return os.WriteFile(file, data, 0644)
}
