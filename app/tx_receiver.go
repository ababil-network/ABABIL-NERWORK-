package app

import (
	"fmt"
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

	if NodeMempool == nil {
		LogError("mempool is not initialized")
		return
	}

	if err := NodeMempool.AdmitTransaction(tx); err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("Added To Mempool")
	LogInfo("Pending Transactions : " + fmt.Sprintf("%d", NodeMempool.Count()))

	LogInfo("=================================")
	LogInfo("Transaction Received")
	LogInfo("Hash : " + tx.Hash)
	LogInfo("=================================")
}
