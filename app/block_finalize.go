package app

func FinalizeBlock(
    validator string,
    block uint64,
    fee uint64,
    mature bool,
) {

    DistributeReward(
        validator,
        block,
        fee,
        mature,
    )
}
