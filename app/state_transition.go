package app

func ApplyTransaction(tx Transaction) error {

	// Validate Transaction
	if err := ValidateTransaction(tx); err != nil {
		return err
	}

	// Sender Balance
	if err := DebitBalance(tx.From, tx.Amount+tx.Fee); err != nil {
		return err
	}

	// Receiver Balance
	CreditBalance(tx.To, tx.Amount)

	// Reward Distribution
	DistributeReward(
		GetLeader().Address,
		0,
		tx.Fee,
		false,
	)

	return nil
}
