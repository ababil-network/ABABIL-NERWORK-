package app

func ApplyTransaction(tx Transaction) error {
	// Validate the complete transaction before changing any state.
	if err := ValidateTransaction(tx); err != nil {
		return err
	}

	// Atomically reserve the transaction hash before state mutation.
	if err := NodeReplay.TryAdd(tx.Hash); err != nil {
		return err
	}

	rollbackReplay := true
	defer func() {
		if rollbackReplay {
			NodeReplay.Remove(tx.Hash)
		}
	}()

	// Serialize transactions from the same sender.
	senderLock := lockSender(tx.From)
	defer unlockSender(senderLock)

	// Atomically reserve the exact next nonce.
	if !NodeNonce.TrySet(tx.From, tx.Nonce) {
		return ErrInvalidNonce
	}

	nonceReserved := true
	defer func() {
		if nonceReserved {
			NodeNonce.Rollback(tx.From, tx.Nonce)
		}
	}()
	// Prevent uint64 overflow before calculating the total debit.
	if tx.Fee > ^uint64(0)-tx.Amount {
		return ErrTransactionValueOverflow
	}

	total := tx.Amount + tx.Fee

	// TransferBalance must atomically debit the sender
	// and credit the receiver.
	if err := TransferBalance(
		tx.From,
		tx.To,
		total,
		tx.Amount,
	); err != nil {
		return err
	}

	// Reward distribution is part of the transaction state transition.
	if tx.Fee > 0 {
		leader := GetLeader()

		if leader != nil {
			if err := DistributeReward(
				leader.Address,
				0,
				tx.Fee,
				false,
			); err != nil {
				// Restore the balance mutation before returning.
				if rollbackErr := TransferBalance(
					tx.To,
					tx.From,
					tx.Amount,
					total,
				); rollbackErr != nil {
					return rollbackErr
				}

				return err
			}
		}
	}

	// Everything succeeded.
	nonceReserved = false
	rollbackReplay = false

	return nil
}
