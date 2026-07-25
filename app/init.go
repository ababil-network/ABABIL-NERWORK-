package app

import (
	"fmt"
       	"os"
	"path/filepath"
)

func InitNode() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	nodeHome := filepath.Join(home, ".ababil")

	dirs := []string{
		filepath.Join(nodeHome, "config"),
		filepath.Join(nodeHome, "data"),
		filepath.Join(nodeHome, "keys"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	fmt.Println("ABABIL node initialized successfully.")
	fmt.Println("Home:", nodeHome)

	return nil
}
