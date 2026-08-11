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
	LogInfo("Initializing node identity...")

	if err := InitNodeKey(); err != nil {
		LogError(err.Error())
		return
	}

	fmt.Println("Node Identity : OK")

	LogInfo("Initializing database...")
	if err := InitDatabase(); err != nil {
		LogError("Failed to initialize database")
		return
	}

	fmt.Println("Database : OK")

	// Restore persistent wallet state after database initialization.
	// A missing state file is valid for a fresh node.
	if state, err := LoadPersistentState(); err == nil {
		if err := ApplyPersistentState(state); err != nil {
			LogError("Failed to apply persistent state: " + err.Error())
			return
		}

		fmt.Printf("Persistent State : LOADED (height=%d)\\n", state.Height)
	} else if os.IsNotExist(err) {
		fmt.Println("Persistent State : NEW")
	} else {
		LogError("Failed to load persistent state: " + err.Error())
		return
	}

	// Restore persistent nonce state.
	// A missing nonce state is valid for a fresh node.
	if nonceState, err := LoadPersistentNonceState(); err == nil {
		if err := ApplyPersistentNonceState(nonceState); err != nil {
			LogError("Failed to apply persistent nonce state: " + err.Error())
			return
		}

		fmt.Println("Nonce State : LOADED")
	} else if os.IsNotExist(err) {
		fmt.Println("Nonce State : NEW")
	} else {
		LogError("Failed to load nonce state: " + err.Error())
		return
	}

	InitMempool()

	fmt.Println("Mempool : OK")

	block := CreateGenesisBlock()
	if err := SaveBlock(block); err != nil {
		LogError("Failed to save genesis block")
		return
	}

	fmt.Println("RPC Server : tcp://0.0.0.0:26657")
	if err := StartP2PServer(); err != nil {
		LogError("Failed to start P2P server")
		return
	}
	fmt.Println("P2P Server : tcp://0.0.0.0:26656")

	if err := InitNodeWallet(); err != nil {
		LogError(err.Error())
		return
	}

	TestTransaction()
	TestWallet()
	TestSignature()
	TestAccount()
	TestBalance()
	TestTransfer()
	TestMempool()
	TestBlockProducer()
	LogInfo("Before TestConsensus")
	TestConsensus()
	TestPeerManager()
	TestP2PConnect()
	TestTxBroadcast()

	fmt.Println("Node Status : Running")
	fmt.Println("=================================")
}
