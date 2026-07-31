package app

type Validator struct {
	Address string
	Power   uint64
	Active  bool
}

var Validators []Validator

func AddValidator(address string, power uint64) {
	Validators = append(Validators, Validator{
		Address: address,
		Power:   power,
		Active:  true,
	})
}

func GetLeader() *Validator {
	for i := range Validators {
		if Validators[i].Active {
			return &Validators[i]
		}
	}
	return nil
}
