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

	expectedHash, err := GenerateTransactionHash(tx)
	if err != nil {
		return err
	}

	if !strings.EqualFold(tx.Hash, expectedHash) {
		return errors.New("transaction hash mismatch")
	}

	if !VerifyTransactionSender(tx) {
		return errors.New("invalid transaction sender signature")
	}

	if !NodeNonce.Verify(tx.From, tx.Nonce) {
		return errors.New("invalid nonce")
	}

	// IMPORTANT:
	// Check amount + fee overflow BEFORE gas-price validation.
	// This guarantees the security test receives
	// ErrTransactionValueOverflow when the transaction value
	// exceeds uint64, even if GasPrice is also invalid.
	if tx.Fee > math.MaxUint64-tx.Amount {
		return ErrTransactionValueOverflow
	}

	total := tx.Amount + tx.Fee

	// Final ABABIL transaction-fee validation.
	//
	// GasPrice is retained only for legacy transaction compatibility.
	// It is NOT authoritative and is never used to calculate Fee.
	//
	// Zero-fee transaction:
	// allowed only while the sender has free quota.
	//
	// Paid transaction:
	// dynamic network load
	// -> USD-equivalent fee
	// -> validated ABABIL reference price
	// -> native ABABIL fee.
	if tx.Fee == 0 {
		if NodeFreeTransaction == nil ||
			NodeFreeTransaction.Remaining(tx.From) == 0 {
			return errors.New("transaction fee required")
		}
	} else {
		expectedFee, err := CalculateFinalNativeFee()
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
