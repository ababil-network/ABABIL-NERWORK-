package app

func TestConsensus() {
        LogInfo("Entering TestConsensus")

	RegisterValidator(
		"0xValidator111111111111111111111111111111",
		100,
	)

	RegisterValidator(
		"0xValidator222222222222222222222222222222",
		80,
	)

	RegisterValidator(
		"0xValidator333333333333333333333333333333",
		60,
	)

	leader := GetLeader()

	if leader != nil {
		LogInfo("=================================")
		LogInfo("Consensus Test")
		LogInfo("=================================")
		LogInfo("Leader Validator")
		LogInfo("Address : " + leader.Address)
	}
}
