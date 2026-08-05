package app

type NodeKey struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NodeID     string `json:"node_id"`
}

var LocalNodeKey *NodeKey
