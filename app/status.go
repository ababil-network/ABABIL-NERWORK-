package app

import "fmt"

func NodeStatus() {
	fmt.Println("========== ABABIL NODE ==========")
	fmt.Println("Network :", AppName)
	fmt.Println("Version :", AppVersion)
	fmt.Println("ChainID :", ChainID)
	fmt.Println("Status  : Ready")
	fmt.Println("=================================")
}
