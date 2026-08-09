package app

func ApplyTransaction(tx Transaction) error {
	// Validate transaction before changing any state.
	if err := ValidateTransaction(tx); err != nil {
		return err
	}

	// Atomically reserve the transaction hash before state mutation.
	if err := NodeReplay.TryAdd(tx.Hash); err != nil {
		return err
	}

	// If execution fails after replay reservation, release the reservation.
	rollbackReplay := true
	defer func() {
		if rollbackReplay {
			NodeReplay.Remove(tx.Hash)
		}
	}()

	total := tx.Amount + tx.Fee

	// Consume free-transaction quota only for zero-fee transactions.
	if tx.Fee == 0 {
		if !NodeFreeTransaction.Use(tx.From) {
			return errGasFeeRequired
		}
	}

	// Sender balance.
	if err := DebitBalance(tx.From, total); err != nil {
		if tx.Fee == 0 {
			// Restore the quota because the transaction did not execute.
			NodeFreeTransaction.Rollback(tx.From)
		}
		return err
	}

	// Receiver balance.
	CreditBalance(tx.To, tx.Amount)

	// Reward distribution.
	if tx.Fee > 0 {
		leader := GetLeader()

		if leader != nil {
			DistributeReward(
				leader.Address,
				0,
				tx.Fee,
				false,
			)
		}
	}

	// Update nonce only after successful state transition.
	NodeNonce.Set(tx.From, tx.Nonce)

	// State transition completed successfully.
	rollbackReplay = false

	return nil
}
