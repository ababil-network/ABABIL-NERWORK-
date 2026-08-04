package app

func Transfer(from *Account, to *Account, amount uint64) error {

	if err := SubBalance(from, amount); err != nil {
		return err
	}

	if err := AddBalance(to, amount); err != nil {
		return err
	}

	return nil
}
