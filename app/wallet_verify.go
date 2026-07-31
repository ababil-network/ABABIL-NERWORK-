package app

type WalletType int

const (
	NormalWallet WalletType = iota
	VerifiedWallet
	SuspiciousWallet
)

type WalletInfo struct {
	Address string
	Type    WalletType
}

var WalletRegistry = map[string]*WalletInfo{}

func RegisterWallet(address string) {
	WalletRegistry[address] = &WalletInfo{
		Address: address,
		Type:    NormalWallet,
	}
}

func VerifyWallet(address string) {
	if w, ok := WalletRegistry[address]; ok {
		w.Type = VerifiedWallet
	}
}

func MarkSuspicious(address string) {
	if w, ok := WalletRegistry[address]; ok {
		w.Type = SuspiciousWallet
	}
}

func WalletFreeLimit(address string) uint64 {
	w, ok := WalletRegistry[address]
	if !ok {
		return 100
	}

	switch w.Type {
	case VerifiedWallet:
		return 500
	case SuspiciousWallet:
		return 0
	default:
		return 100
	}
}
