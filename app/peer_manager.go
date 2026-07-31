package app

type Peer struct {
	Address string
	Active  bool
}

var Peers []Peer

func AddPeer(address string) {
	for _, p := range Peers {
		if p.Address == address {
			return
		}
	}

	Peers = append(Peers, Peer{
		Address: address,
		Active:  true,
	})
}

func RemovePeer(address string) {
	var newPeers []Peer

	for _, p := range Peers {
		if p.Address != address {
			newPeers = append(newPeers, p)
		}
	}

	Peers = newPeers
}

func GetPeers() []Peer {
	return Peers
}

func PeerCount() int {
	return len(Peers)
}
