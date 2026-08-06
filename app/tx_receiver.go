package app

import (
	"io"
	"net"
)

func HandleTransaction(conn net.Conn) {

	defer conn.Close()

	var tx Transaction

	err := ReceiveJSON(conn, &tx)
	if err != nil {
		if err == io.EOF {
			return
		}
		LogError(err.Error())
		return
	}

	err = ValidateTransaction(tx)
	if err != nil {
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
