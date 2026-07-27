package app

type Account struct {
	Address string
	Balance uint64
}

func NewAccount(address string, balance uint64) Account {
	return Account{
		Address: address,
		Balance: balance,
	}
}
