package app

import (
	"encoding/json"
	"io"
	"net"
	"strconv"
)

func HandleBlock(conn net.Conn) {

	defer conn.Close()

	var block Block

	err := json.NewDecoder(conn).Decode(&block)
	if err != nil {
		if err == io.EOF {
			return
		}
		LogError(err.Error())
		return
	}

	LogInfo("=================================")
	LogInfo("Block Received")
	LogInfo("Height : " + strconv.Itoa(block.Height))
	LogInfo("Hash : " + block.Hash)
	LogInfo("=================================")
latest, err := GetLatestBlock()
if err != nil {
	LogError(err.Error())
	return
}

err = ValidateBlock(block, latest)
if err != nil {
	LogError(err.Error())
	return
}

CommitBlock(block)

LogInfo("Block Successfully Added")

}
