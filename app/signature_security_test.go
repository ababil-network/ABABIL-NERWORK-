package app

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

func TestSignTransactionWithPrivateKeyProducesValidSignature(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatalf("CreateWallet failed: %v", err)
	}

	hash := crypto.Keccak256Hash([]byte("ABABIL signature test")).Hex()[2:]

	signed := SignTransactionWithPrivateKey(hash, wallet.PrivateKey)

	if signed.Hash != hash {
		t.Fatalf("hash mismatch: got %q want %q", signed.Hash, hash)
	}

	if signed.Signature == "" {
		t.Fatal("signature is empty")
	}

	if signed.PublicKey == "" {
		t.Fatal("public key is empty")
	}

	if !VerifyTransaction(signed) {
		t.Fatal("valid signature failed verification")
	}
}

func TestVerifyTransactionRejectsTamperedHash(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256Hash([]byte("original")).Hex()[2:]
	signed := SignTransactionWithPrivateKey(hash, wallet.PrivateKey)

	tampered := signed
	tampered.Hash = crypto.Keccak256Hash([]byte("tampered")).Hex()[2:]

	if VerifyTransaction(tampered) {
		t.Fatal("tampered hash was accepted")
	}
}

func TestVerifyTransactionRejectsTamperedSignature(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256Hash([]byte("signature test")).Hex()[2:]
	signed := SignTransactionWithPrivateKey(hash, wallet.PrivateKey)

	raw, err := hex.DecodeString(signed.Signature)
	if err != nil {
		t.Fatal(err)
	}

	raw[0] ^= 0x01
	tampered := signed
	tampered.Signature = hex.EncodeToString(raw)

	if VerifyTransaction(tampered) {
		t.Fatal("tampered signature was accepted")
	}
}

func TestVerifyTransactionRejectsWrongPublicKey(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	other, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256Hash([]byte("public key test")).Hex()[2:]
	signed := SignTransactionWithPrivateKey(hash, wallet.PrivateKey)

	tampered := signed
	tampered.PublicKey = other.PublicKey

	if VerifyTransaction(tampered) {
		t.Fatal("wrong public key was accepted")
	}
}

func TestVerifyTransactionSenderRejectsWrongSender(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	other, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Keccak256Hash([]byte("sender test")).Hex()[2:]
	signed := SignTransactionWithPrivateKey(hash, wallet.PrivateKey)

	tx := Transaction{
		From:      other.Address,
		Hash:      signed.Hash,
		Signature: signed.Signature,
		PublicKey: signed.PublicKey,
	}

	if VerifyTransactionSender(tx) {
		t.Fatal("signature was accepted for wrong sender")
	}
}

func TestSignTransactionWithPrivateKeyRejectsInvalidInput(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		hash string
		key  string
	}{
		{
			name: "empty hash",
			hash: "",
			key:  wallet.PrivateKey,
		},
		{
			name: "empty private key",
			hash: crypto.Keccak256Hash([]byte("test")).Hex()[2:],
			key:  "",
		},
		{
			name: "invalid hash",
			hash: "not-a-hash",
			key:  wallet.PrivateKey,
		},
		{
			name: "short hash",
			hash: "abcd",
			key:  wallet.PrivateKey,
		},
		{
			name: "invalid private key",
			hash: crypto.Keccak256Hash([]byte("test")).Hex()[2:],
			key:  "invalid-private-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signed := SignTransactionWithPrivateKey(tt.hash, tt.key)

			if signed.Signature != "" || signed.PublicKey != "" {
				t.Fatal("invalid input unexpectedly produced a signature")
			}
		})
	}
}

func TestValidatePrivateKeyForAddress(t *testing.T) {
	wallet, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidatePrivateKeyForAddress(
		wallet.PrivateKey,
		wallet.Address,
	); err != nil {
		t.Fatalf("valid private key/address pair rejected: %v", err)
	}

	other, err := CreateWallet()
	if err != nil {
		t.Fatal(err)
	}

	if err := ValidatePrivateKeyForAddress(
		wallet.PrivateKey,
		other.Address,
	); err == nil {
		t.Fatal("private key was accepted for another address")
	}

	if err := ValidatePrivateKeyForAddress(
		"",
		wallet.Address,
	); err == nil {
		t.Fatal("empty private key was accepted")
	}

	if err := ValidatePrivateKeyForAddress(
		wallet.PrivateKey,
		"invalid-address",
	); err == nil {
		t.Fatal("invalid address was accepted")
	}
}
