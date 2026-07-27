package app

func TestAccount() {
	account := NewAccount(
		"0xABA4F2D81C8A97E31B5A8F4D0E91C2A7B84E1D3",
		1000,
	)

	if err := SaveAccount(account); err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("Account created successfully")
	LogInfo("Address : " + account.Address)
}
