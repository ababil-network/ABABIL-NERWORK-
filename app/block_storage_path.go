package app

import (
	"os"
	"path/filepath"
)

var blockStorageRoot string

func BlockStorageDir() (string, error) {
	if blockStorageRoot != "" {
		return filepath.Join(blockStorageRoot, "blocks"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(
		home,
		".ababil",
		"data",
		"blocks",
	), nil
}
