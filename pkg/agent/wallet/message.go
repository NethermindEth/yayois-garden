package wallet

import (
	"fmt"
	"log/slog"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	beecrypto "github.com/ethersphere/bee/pkg/crypto"
)

var eip712Types = apitypes.Types{
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	},
	"Mint": []apitypes.Type{
		{Name: "to", Type: "address"},
		{Name: "uri", Type: "string"},
	},
}

type EIP712Domain struct {
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
}

func (w *Wallet) SignMintMessage(to common.Address, uri string, domain EIP712Domain) ([]byte, error) {
	signer := beecrypto.NewDefaultSigner(w.privateKey)

	address, err := signer.EthereumAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get ethereum address: %w", err)
	}
	slog.Info("signer", "signer", address.Hex())

	signature, err := signer.SignTypedData(&apitypes.TypedData{
		Types:       eip712Types,
		PrimaryType: "Mint",
		Domain: apitypes.TypedDataDomain{
			Name:              domain.Name,
			Version:           domain.Version,
			ChainId:           (*math.HexOrDecimal256)(domain.ChainId),
			VerifyingContract: domain.VerifyingContract.Hex(),
		},
		Message: map[string]interface{}{
			"to":  to.Hex(),
			"uri": uri,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign mint message: %w", err)
	}

	return signature, nil
}
