package app

func NeedSync(localHeight int, peerHeight int) bool {
	return peerHeight > localHeight
}
