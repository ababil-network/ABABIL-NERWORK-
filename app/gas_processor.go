package app

import "fmt"

func ChargeGas(balance uint64, fee uint64) (uint64, error) {

	if balance < fee {
		return balance, fmt.Errorf("insufficient balance for gas fee")
	}

	return balance - fee, nil
}
