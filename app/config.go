package app

import (
	"fmt"
	"os"
)

func LoadConfig() error {
	_, err := os.ReadFile("config/config.toml")
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Configuration loaded successfully.")
	return nil
}
