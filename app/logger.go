package app

import "fmt"

func LogInfo(message string) {
	fmt.Println("[INFO]", message)
}

func LogError(message string) {
	fmt.Println("[ERROR]", message)
}
