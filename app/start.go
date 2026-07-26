package app

import (
        "fmt"
        "os"
)

func StartNode() {
        fmt.Println("=================================")
        fmt.Println("        ABABIL NETWORK")
        fmt.Println("=================================")

        fmt.Println("Starting ABABIL Node...")

        fmt.Println("Loading configuration...")
        if _, err := os.Stat("config/config.toml"); err != nil {
                fmt.Println("Config : NOT FOUND")
        } else {
                fmt.Println("Config : OK")
        }

        fmt.Println("Loading genesis...")
        if _, err := os.Stat("genesis/genesis.json"); err != nil {
                fmt.Println("Genesis : NOT FOUND")
        } else {
                fmt.Println("Genesis : OK")
        }

        fmt.Println("Initializing database...")
        fmt.Println("RPC Server : tcp://0.0.0.0:26657")
        fmt.Println("P2P Server : tcp://0.0.0.0:26656")
        fmt.Println("Node Status : Running")
        fmt.Println("=================================")
}