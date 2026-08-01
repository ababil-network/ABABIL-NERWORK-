package app

import (
	"encoding/json"
	"io"
	"net"
)

func HandleTransaction(conn net.Conn) {

	defer conn.Close()

	var tx Transaction

	err := json.NewDecoder(conn).Decode(&tx)
	if err != nil {
		if err == io.EOF {
		return
	}
	LogError(err.Error())
		return
	}

mempool := NewMempool()
mempool.AddTransaction(tx)

LogInfo("Added To Mempool")
LogInfo("Pending Transactions : 1")

	LogInfo("=================================")
	LogInfo("Transaction Received")
	LogInfo("Hash : " + tx.Hash)
	LogInfo("=================================")
}
