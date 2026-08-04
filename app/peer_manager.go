package app

import "sync"

type Peer struct {
	Address string
	Active  bool
}

var PeerMutex sync.RWMutex
var Peers []Peer

func AddPeer(address string) {
	PeerMutex.Lock()
	defer PeerMutex.Unlock()

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
	PeerMutex.Lock()
	defer PeerMutex.Unlock()

	var newPeers []Peer

	for _, p := range Peers {
		if p.Address != address {
			newPeers = append(newPeers, p)
		}
	}

	Peers = newPeers
}

func GetPeers() []Peer {
	PeerMutex.RLock()
	defer PeerMutex.RUnlock()

	return Peers
}

func PeerCount() int {
	PeerMutex.RLock()
	defer PeerMutex.RUnlock()

	return len(Peers)
}
func HasPeer(address string) bool {
	PeerMutex.RLock()
	defer PeerMutex.RUnlock()

	for _, p := range Peers {

		if p.Address == address && p.Active {
			return true
		}
	}

	return false
}
