
}
package main

import (
	"fmt"
	"os"

	"github.com/ababil-network/ababil/app"
)

func main() {
	if len(os.Args) < 2 {
		printBanner()
		return
	}

	switch os.Args[1] {
	case "version":
		printBanner()

	case "init":
		fmt.Println("Initializing ABABIL node...")
		fmt.Println("Chain ID:", app.ChainID)
		fmt.Println("Status: Coming Soon")

	case "start":
		fmt.Println("Starting ABABIL node...")
		fmt.Println("Status: Coming Soon")

	case "status":
		fmt.Println("ABABIL node status")
		fmt.Println("Status: Offline")

	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Available commands:")
		fmt.Println("  version")
		fmt.Println("  init")
		fmt.Println("  start")
		fmt.Println("  status")
	}
}

func printBanner() {
	fmt.Println("=================================")
	fmt.Println("        ABABIL NETWORK")
	fmt.Println("=================================")
	fmt.Println("Version :", app.Version)
	fmt.Println("Chain ID:", app.ChainID)
	fmt.Println("Coin    :", app.Denom)
	fmt.Println("Binary  :", app.BinaryName)
	fmt.Println("=================================")
}
