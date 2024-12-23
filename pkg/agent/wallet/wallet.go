package wallet

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	seed       []byte
}

func NewWallet(seed []byte) (*Wallet, error) {
	privateKey, err := crypto.ToECDSA(crypto.Keccak256(seed))
	if err != nil {
		return nil, err
	}

	return &Wallet{
		privateKey: privateKey,
		seed:       seed,
	}, nil
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) Address() common.Address {
	return crypto.PubkeyToAddress(w.privateKey.PublicKey)
}

func (w *Wallet) Sign(data []byte) ([]byte, error) {
	return crypto.Sign(data, w.privateKey)
}
