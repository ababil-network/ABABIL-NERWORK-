package app

import (
	"fmt"
	"os"
	"path/filepath"
)

func InitDatabase() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dbPath := filepath.Join(home, ".ababil", "data")

	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return err
	}

	fmt.Println("[INFO] Database initialized successfully.")
	fmt.Println("[INFO] Database Path:", dbPath)

	return nil
}
