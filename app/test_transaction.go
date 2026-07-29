package app

import "fmt"

func TestTransaction() {

	tx := NewTransaction(
		NodeWallet.Address,
		"0x1111111111111111111111111111111111111111",
		100,
	)

	LogInfo("=================================")
	LogInfo("Transaction Created")
	LogInfo("=================================")
	LogInfo("Hash      : " + tx.Hash)
	LogInfo("Signature : " + tx.Signature)

	LogInfo(fmt.Sprintf("Gas Limit : %d", tx.GasLimit))
	LogInfo(fmt.Sprintf("Gas Price : %d", tx.GasPrice))
	LogInfo(fmt.Sprintf("Gas Fee   : %d", tx.Fee))

	balance := uint64(100000)

	LogInfo(fmt.Sprintf("Balance Before : %d", balance))

	newBalance, err := ChargeGas(balance, tx.Fee)
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo(fmt.Sprintf("Balance After  : %d", newBalance))

	err = ValidateTransaction(tx)
if err != nil {
        LogError("Transaction Verify : " + err.Error())
        return
}

LogInfo("Transaction Verify : VALID")
}
