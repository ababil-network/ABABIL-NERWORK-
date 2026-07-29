package app

func TestWallet() {

	if NodeWallet == nil {
		LogError("Node Wallet not initialized")
		return
	}

	LogInfo("=================================")
	LogInfo("ABABIL Production Wallet")
	LogInfo("=================================")
	LogInfo("Network    : ABABIL Network")
	LogInfo("Chain ID   : 7777")
	LogInfo("Address    : " + NodeWallet.Address)
	LogInfo("PrivateKey : " + NodeWallet.PrivateKey)
	LogInfo("PublicKey  : " + NodeWallet.PublicKey)
	LogInfo("=================================")
}
