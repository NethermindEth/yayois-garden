package setup

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/NethermindEth/yayois-garden/pkg/agent/debug"
	"github.com/NethermindEth/yayois-garden/pkg/agent/sealing"
)

type SetupResult struct {
	DstackTappdEndpoint   string
	EthereumRpcUrl        string
	FactoryAddress        common.Address
	SecureFile            string
	OpenAiApiKey          string
	OpenAiModel           string
	PinataJwtKey          string
	ApiIpPort             string
	AccountPrivateKeySeed []byte
	RsaPrivateKey         *rsa.PrivateKey
}

func Setup(ctx context.Context) (*SetupResult, error) {
	config, err := NewConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get config from env: %v", err)
	}

	setupResult, err := loadSetup(ctx, config)
	if err != nil {
		slog.Warn("failed to load setup, initializing new setup", "error", err)
		return initializeSetup(ctx, config)
	}

	if debug.IsDebugShowSetup() {
		slog.Info("setup output", "setupOutput", setupResult)
	}

	return setupResult, nil
}

func generateSetup(config *Config) (*SetupResult, error) {
	accountPrivateKeySeed := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, accountPrivateKeySeed); err != nil {
		return nil, fmt.Errorf("failed to generate private key seed: %v", err)
	}

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rsa private key: %w", err)
	}

	return &SetupResult{
		DstackTappdEndpoint:   config.DstackTappdEndpoint,
		EthereumRpcUrl:        config.EthereumRpcUrl,
		FactoryAddress:        common.HexToAddress(config.FactoryAddress),
		SecureFile:            config.SecureFile,
		OpenAiApiKey:          config.OpenAiApiKey,
		OpenAiModel:           config.OpenAiModel,
		PinataJwtKey:          config.PinataJwtKey,
		ApiIpPort:             config.ApiIpPort,
		AccountPrivateKeySeed: accountPrivateKeySeed,
		RsaPrivateKey:         rsaPrivateKey,
	}, nil
}

func initializeSetup(ctx context.Context, config *Config) (*SetupResult, error) {
	setupResult, err := generateSetup(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate setup: %v", err)
	}

	if err := writeSetupResult(ctx, config, setupResult); err != nil {
		return nil, fmt.Errorf("failed to write setup output: %v", err)
	}

	slog.Info("wrote encrypted setup output")

	return setupResult, nil
}

func loadSetup(ctx context.Context, config *Config) (*SetupResult, error) {
	setupResult, err := readSetupResult(ctx, config)
	if err != nil {
		return nil, err
	}

	slog.Info("loaded decrypted setup output")

	return setupResult, nil
}

func writeSetupResult(ctx context.Context, config *Config, setupResult *SetupResult) error {
	data, err := json.Marshal(setupResult)
	if err != nil {
		return fmt.Errorf("failed to marshal setup result: %v", err)
	}

	return sealing.WriteSealedFile(ctx, config.DstackTappdEndpoint, config.SecureFile, data)
}

func readSetupResult(ctx context.Context, config *Config) (*SetupResult, error) {
	data, err := sealing.ReadSealedFile(ctx, config.DstackTappdEndpoint, config.SecureFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read sealed file: %v", err)
	}

	var setupResult SetupResult
	if err := json.Unmarshal(data, &setupResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal setup result: %v", err)
	}

	return &setupResult, nil
}
