package app

type Handshake struct {
	ProtocolVersion uint32 `json:"protocol_version"`
	ChainID         uint64 `json:"chain_id"`
	Network         string `json:"network"`
	NodeName        string `json:"node_name"`
	NodeVersion     string `json:"node_version"`
}
