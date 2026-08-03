package app

import "errors"

func ValidateTransaction(tx Transaction) error {

	// Replay Protection
	if !NodeReplay.Check(tx.Hash) {
		return errors.New("replay transaction detected")
	}

	// Signature Verification
	signed := SignedTransaction{
		Hash:      tx.Hash,
		Signature: tx.Signature,
		PublicKey: tx.PublicKey,
	}

	if !VerifyTransaction(signed) {
		return errors.New("invalid signature")
	}

	// Nonce Verification
	if !NodeNonce.Verify(tx.From, tx.Nonce) {
		return errors.New("invalid nonce")
	}

	// Amount
	if tx.Amount == 0 {
		return errors.New("amount must be greater than zero")
	}

	// Daily Free Transactions
	if NodeFreeTransaction.Use(tx.From) {
		tx.Fee = 0
	} else {
		if tx.Fee == 0 {
			return errors.New("gas fee required")
		}
	}
	// Gas
	if tx.GasLimit == 0 {
		return errors.New("invalid gas limit")
	}

	// Gas Price (only check if fee is required)
	if tx.Fee > 0 && tx.GasPrice == 0 {
		return errors.New("invalid gas price")
	}

        // Balance Verification
        if GetBalance(tx.From) < (tx.Amount + tx.Fee) {
         return errors.New("insufficient balance")
}
	// Success
	NodeReplay.Add(tx.Hash)
	NodeNonce.Set(tx.From, tx.Nonce)

	return nil
}
