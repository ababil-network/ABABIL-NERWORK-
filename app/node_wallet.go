package app

var NodeWallet *Wallet

func InitNodeWallet() error {

	wallet, err := LoadWallet()
	if err == nil {
		NodeWallet = wallet
		LogInfo("Existing wallet loaded.")
		return nil
	}

	wallet, err = CreateWallet()
	if err != nil {
		return err
	}

	if err := SaveWallet(wallet); err != nil {
		return err
	}

	NodeWallet = wallet

	LogInfo("New wallet created and saved.")

	return nil
}
