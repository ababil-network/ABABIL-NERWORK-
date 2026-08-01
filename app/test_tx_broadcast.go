package app

import (
    "time"
)

func TestTxBroadcast() {

	LogInfo("=================================")
	LogInfo("Transaction Broadcast Test")
	LogInfo("=================================")

	tx := NewTransaction(
		"0xABA1111111111111111111111111111111111111",
		"0xABA2222222222222222222222222222222222222",
		50,
	)

	err := BroadcastTransaction(tx, "127.0.0.1:26656")
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("Transaction Broadcast Success")
        time.Sleep(2 * time.Second)
}
