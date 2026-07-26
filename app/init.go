package app

import (
	"fmt"
	"io"
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
	fmt.Println("Config :", filepath.Join(nodeHome, "config"))
	fmt.Println("Data   :", filepath.Join(nodeHome, "data"))
	fmt.Println("Keys   :", filepath.Join(nodeHome, "keys"))

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}