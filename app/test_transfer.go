package app

func TestTransfer() {

	from := NewAccount(
		"0xABA1111111111111111111111111111111111111",
		1200,
	)

	to := NewAccount(
		"0xABA2222222222222222222222222222222222222",
		300,
	)

	LogInfo("Before Transfer")
	LogInfo("Sender Balance : 1200 ABABIL")
	LogInfo("Receiver Balance : 300 ABABIL")

	err := Transfer(&from, &to, 500)
	if err != nil {
		LogError(err.Error())
		return
	}

	LogInfo("Transfer Success")
	LogInfo("Sender Balance : 700 ABABIL")
	LogInfo("Receiver Balance : 800 ABABIL")
}
