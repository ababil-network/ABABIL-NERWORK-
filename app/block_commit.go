package app

var Blockchain []Block

func CommitBlock(block Block) {

	Blockchain = append(Blockchain, block)

	LogInfo("=================================")
	LogInfo("Block Committed")
	LogInfo("Block Height Recorded")
	LogInfo("Hash   : " + block.Hash)
	LogInfo("=================================")
} 
