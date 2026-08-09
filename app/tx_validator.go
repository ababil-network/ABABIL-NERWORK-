package app

import (
	"errors"
	"math"
	"strings"
)

func ValidateTransaction(tx Transaction) error {
	if tx.ID == "" {
		return errors.New("transaction ID is empty")
	}

	if !IsValidAddress(tx.From) {
		return errors.New("invalid sender address")
	}

	if !IsValidAddress(tx.To) {
		return errors.New("invalid receiver address")
	}

	if strings.EqualFold(tx.From, tx.To) {
		return errors.New("sender and receiver cannot be the same")
	}

	if tx.Amount == 0 {
		return errors.New("amount must be greater than zero")
	}

	if tx.GasLimit == 0 {
		return errors.New("invalid gas limit")
	}

	if tx.GasLimit > MaxGasLimit {
		return errors.New("gas limit exceeds maximum")
	}

	if tx.Timestamp.IsZero() {
		return errors.New("transaction timestamp is empty")
	}

	// Rebuild canonical transaction hash.
	expectedHash, err := GenerateTransactionHash(tx)
	if err != nil {
		return err
	}

	if !strings.EqualFold(tx.Hash, expectedHash) {
		return errors.New("transaction hash mismatch")
	}

	// Verify signature and ensure the signing key owns tx.From.
	if !VerifyTransactionSender(tx) {
		return errors.New("invalid transaction sender signature")
	}

	// Nonce must be exactly the next nonce.
	if !NodeNonce.Verify(tx.From, tx.Nonce) {
		return errors.New("invalid nonce")
	}

	// Prevent uint64 overflow in amount + fee.
	if tx.Fee > math.MaxUint64-tx.Amount {
		return errors.New("transaction value overflow")
	}

	total := tx.Amount + tx.Fee

	// Zero fee is valid only while the free-transaction quota is available.
	if tx.Fee == 0 {
		if NodeFreeTransaction.Remaining(tx.From) == 0 {
			return errors.New("gas fee required")
		}
	} else {
		if tx.GasPrice == 0 {
			return errors.New("invalid gas price")
		}

		expectedFee, err := CalculateGasFee(tx.GasLimit, tx.GasPrice)
		if err != nil {
			return err
		}

		if tx.Fee != expectedFee {
			return errors.New("invalid transaction fee")
		}
	}

	if GetBalance(tx.From) < total {
		return errors.New("insufficient balance")
	}

	return nil
}
