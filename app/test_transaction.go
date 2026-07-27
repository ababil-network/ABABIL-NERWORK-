package app

func TestTransaction() {
	tx := NewTransaction(
		"ABABIL1FROM",
		"ABABIL1TO",
		100,
	)

	if err := SaveTransaction(tx); err != nil {
		LogError(err.Error())
	}
}
