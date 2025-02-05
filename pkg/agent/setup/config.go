package setup

import (
	"errors"
	"os"
)

type Config struct {
	DstackTappdEndpoint string
	EthereumRpcUrl      string
	FactoryAddress      string
	SecureFile          string
	OpenAiApiKey        string
	OpenAiModel         string
	PinataJwtKey        string
	ApiIpPort           string
}

func NewConfigFromEnv() (*Config, error) {
	config := &Config{
		DstackTappdEndpoint: os.Getenv(EnvDstackTappdEndpoint),
		EthereumRpcUrl:      os.Getenv(EnvEthereumRpcUrl),
		FactoryAddress:      os.Getenv(EnvFactoryAddress),
		SecureFile:          os.Getenv(EnvSecureFile),
		OpenAiApiKey:        os.Getenv(EnvOpenAiApiKey),
		OpenAiModel:         os.Getenv(EnvOpenAiModel),
		PinataJwtKey:        os.Getenv(EnvPinataJwtKey),
		ApiIpPort:           os.Getenv(EnvApiIpPort),
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.DstackTappdEndpoint == "" {
		return errors.New(EnvDstackTappdEndpoint + " is required")
	}
	if c.EthereumRpcUrl == "" {
		return errors.New(EnvEthereumRpcUrl + " is required")
	}
	if c.FactoryAddress == "" {
		return errors.New(EnvFactoryAddress + " is required")
	}
	if c.SecureFile == "" {
		return errors.New(EnvSecureFile + " is required")
	}
	if c.OpenAiApiKey == "" {
		return errors.New(EnvOpenAiApiKey + " is required")
	}
	if c.OpenAiModel == "" {
		return errors.New(EnvOpenAiModel + " is required")
	}
	if c.PinataJwtKey == "" {
		return errors.New(EnvPinataJwtKey + " is required")
	}
	if c.ApiIpPort == "" {
		return errors.New(EnvApiIpPort + " is required")
	}
	return nil
}
