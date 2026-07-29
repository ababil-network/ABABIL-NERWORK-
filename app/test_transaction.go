package app

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

	signed := SignedTransaction{
		Hash:      tx.Hash,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
	}

	if VerifyTransaction(signed) {
		LogInfo("Transaction Verify : VALID")
	} else {
		LogError("Transaction Verify : INVALID")
	}
}
