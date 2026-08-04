package app

func TestBalance() {
	account := NewAccount(
		"0xABA4F2D81C8A97E31B5A8F4D0E91C2A7B84E1D3",
		1000,
	)

	LogInfo("Current Balance : 1000 ABABIL")

	if err := AddBalance(&account, 500); err != nil {
		LogError(err.Error())
		return
	}
	LogInfo("After Add : 1500 ABABIL")

	err := SubBalance(&account, 300)
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("After Sub : 1200 ABABIL")
}
