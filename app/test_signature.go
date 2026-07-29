package app

func TestSignature() {

	tx := SignTransaction("TX_HASH_001")

	if VerifyTransaction(tx) {
		LogInfo("Signature Verification : VALID")
	} else {
		LogError("Signature Verification : INVALID")
	}
}
