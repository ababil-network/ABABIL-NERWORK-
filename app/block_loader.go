package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func LoadBlock(height int) (Block, error) {
	dir, err := BlockStorageDir()
	if err != nil {
		return Block{}, err
	}

	file := filepath.Join(
		dir,
		fmt.Sprintf("%d.json", height),
	)

	data, err := os.ReadFile(file)
	if err != nil {
		return Block{}, err
	}

	var block Block

	if err := json.Unmarshal(data, &block); err != nil {
		return Block{}, err
	}

	return block, nil
}

func GetLatestBlock() (Block, error) {
	dir, err := BlockStorageDir()
	if err != nil {
		return Block{}, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return Block{}, err
	}

	var latest Block

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var height int

		_, err := fmt.Sscanf(entry.Name(), "%d.json", &height)
		if err != nil {
			continue
		}

		block, err := LoadBlock(height)
		if err != nil {
			continue
		}

		if block.Height >= latest.Height {
			latest = block
		}
	}

	return latest, nil
}
