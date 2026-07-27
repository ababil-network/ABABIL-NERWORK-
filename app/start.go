package app

import (
        "fmt"
        "os"
)

func StartNode() {
        fmt.Println("=================================")
        fmt.Println("        ABABIL NETWORK")
        fmt.Println("=================================")

        LogInfo("Starting ABABIL Node...")

        fmt.Println("Loading configuration...")
if err := LoadConfig(); err != nil {
        LogError("Failed to load configuration")
        return
}
        if _, err := os.Stat("config/config.toml"); err != nil {
                fmt.Println("Config : NOT FOUND")
        } else {
                fmt.Println("Config : OK")
        }

        fmt.Println("Loading genesis...")
if err := LoadGenesis(); err != nil {
        LogError("Failed to load genesis")
        return
}
        if _, err := os.Stat("genesis/genesis.json"); err != nil {
                fmt.Println("Genesis : NOT FOUND")
        } else {
                fmt.Println("Genesis : OK")
        }

        LogInfo("Initializing database...")
if err := InitDatabase(); err != nil {
        LogError("Failed to initialize database")
        return
}
        fmt.Println("Database : OK")
        block := CreateGenesisBlock()

if err := SaveBlock(block); err != nil {
        LogError("Failed to save genesis block")
        return
}
        fmt.Println("RPC Server : tcp://0.0.0.0:26657")
        fmt.Println("P2P Server : tcp://0.0.0.0:26656")
        TestTransaction()
        TestWallet()
        TestAccount()
        TestBalance()
        TestTransfer()
        TestMempool()
        TestBlockProducer()
        fmt.Println("Node Status : Running")
        fmt.Println("=================================")
}
