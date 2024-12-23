package wallet

import (
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Wallet struct {
	privateKey *ecdsa.PrivateKey
	seed       []byte
	auth       *bind.TransactOpts
}

func NewWallet(seed []byte, chainID *big.Int) (*Wallet, error) {
	privateKey, err := crypto.ToECDSA(crypto.Keccak256(seed))
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		privateKey: privateKey,
		seed:       seed,
		auth:       auth,
	}, nil
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) Address() common.Address {
	return crypto.PubkeyToAddress(w.privateKey.PublicKey)
}

func (w *Wallet) Auth() *bind.TransactOpts {
	return w.auth
}
