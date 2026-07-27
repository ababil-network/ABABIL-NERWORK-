package app

import (
	"fmt"
	"os"
)

func LoadGenesis() error {
	_, err := os.ReadFile("genesis/genesis.json")
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Genesis loaded successfully.")
	return nil
}
