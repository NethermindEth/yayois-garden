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
}

func NewConfigFromEnv() (*Config, error) {
	config := &Config{
		DstackTappdEndpoint: os.Getenv(EnvDstackTappdEndpoint),
		EthereumRpcUrl:      os.Getenv(EnvEthereumRpcUrl),
		FactoryAddress:      os.Getenv(EnvFactoryAddress),
		SecureFile:          os.Getenv(EnvSecureFile),
		OpenAiApiKey:        os.Getenv(EnvOpenAiApiKey),
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.DstackTappdEndpoint == "" {
		return errors.New("DSTACK_TAPPD_ENDPOINT is required")
	}
	if c.EthereumRpcUrl == "" {
		return errors.New("ETHEREUM_RPC_URL is required")
	}
	if c.FactoryAddress == "" {
		return errors.New("FACTORY_ADDRESS is required")
	}
	if c.SecureFile == "" {
		return errors.New("SECURE_FILE is required")
	}
	if c.OpenAiApiKey == "" {
		return errors.New("OPENAI_API_KEY is required")
	}

	return nil
}
