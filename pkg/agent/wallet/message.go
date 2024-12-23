package wallet

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/ethersphere/bee/pkg/crypto/eip712"
)

var eip712Types = apitypes.Types{
	"Mint": []apitypes.Type{
		{Name: "to", Type: "address"},
		{Name: "uri", Type: "string"},
	},
}

func (w *Wallet) MintMessage(to common.Address, uri string, domain apitypes.TypedDataDomain) ([]byte, error) {
	typedData, err := eip712.EncodeForSigning(&apitypes.TypedData{
		Types:       eip712Types,
		PrimaryType: "Mint",
		Domain:      domain,
		Message: map[string]interface{}{
			"to":  to,
			"uri": uri,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode mint message: %w", err)
	}

	return typedData, nil
}

func (w *Wallet) SignMintMessage(to common.Address, uri string, domain apitypes.TypedDataDomain) ([]byte, error) {
	message, err := w.MintMessage(to, uri, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create mint message: %w", err)
	}

	return crypto.Sign(message, w.privateKey)
}
