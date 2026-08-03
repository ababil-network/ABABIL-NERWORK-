package app

import (
	"crypto/ecdsa"
	"encoding/hex"
        
        "github.com/ethereum/go-ethereum/common"
        "github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	PrivateKey string
	PublicKey  string
	Address    string
}

func CreateWallet() (*Wallet, error) {

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	privateBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	publicBytes := crypto.FromECDSAPub(publicKey)

	address := crypto.PubkeyToAddress(*publicKey)

	return &Wallet{
		PrivateKey: hex.EncodeToString(privateBytes),
		PublicKey:  hex.EncodeToString(publicBytes),
		Address:    address.Hex(),
	}, nil
}
func IsValidAddress(address string) bool {
    return common.IsHexAddress(address)
}

func ImportWallet(privateKeyHex string) (*Wallet, error) {

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	privateBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	publicBytes := crypto.FromECDSAPub(publicKey)

	address := crypto.PubkeyToAddress(*publicKey)

	return &Wallet{
		PrivateKey: hex.EncodeToString(privateBytes),
		PublicKey:  hex.EncodeToString(publicBytes),
		Address:    address.Hex(),
	}, nil
}
