package app

func TestAccount() {

	if NodeWallet == nil {
		LogError("Node Wallet not initialized")
		return
	}

	account := NewAccount(
		NodeWallet.Address,
		1000,
	)

	if err := SaveAccount(account); err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("Account created successfully")
	LogInfo("Address : " + account.Address)
}
