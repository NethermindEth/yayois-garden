package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"

	"github.com/NethermindEth/yayois-garden/pkg/agent/art"
	"github.com/NethermindEth/yayois-garden/pkg/agent/filestorage"
	"github.com/NethermindEth/yayois-garden/pkg/agent/indexer"
	"github.com/NethermindEth/yayois-garden/pkg/agent/nft"
	"github.com/NethermindEth/yayois-garden/pkg/agent/setup"
	"github.com/NethermindEth/yayois-garden/pkg/agent/wallet"
	contractYayoiCollection "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiCollection"
)

type AgentEthClient interface {
	bind.ContractBackend
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainIDReader
}

type Agent struct {
	artGenerator art.ArtGenerator
	indexer      *indexer.Indexer
	ethClient    AgentEthClient
	wallet       *wallet.Wallet
	nftUploader  *nft.NftUploader
	tappdClient  TappdClient
	apiRouter    *gin.Engine
	httpClient   *http.Client

	rsaPrivateKey *rsa.PrivateKey

	factoryAddress  common.Address
	pollingInterval time.Duration
	apiIpPort       string

	mu sync.Mutex
}

type AgentConfig struct {
	ArtGenerator art.ArtGenerator
	Uploader     filestorage.Uploader
	EthClient    AgentEthClient
	TappdClient  TappdClient
	HttpClient   *http.Client

	FactoryAddress  common.Address
	PollingInterval time.Duration
	PrivateKeySeed  []byte
	ApiIpPort       string
}

func NewAgent(ctx context.Context, config *AgentConfig) (*Agent, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	rsaPrivateKey, err := rsa.GenerateKey(bytes.NewReader(config.PrivateKeySeed), 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate rsa private key: %w", err)
	}

	indexer, err := indexer.NewIndexer(indexer.IndexerOptions{
		EthClient:       config.EthClient,
		ContractAddress: config.FactoryAddress,
		PollingInterval: config.PollingInterval,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create indexer: %w", err)
	}

	chainID, err := config.EthClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %w", err)
	}

	wallet, err := wallet.NewWallet(config.PrivateKeySeed, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	nftUploader := nft.NewNftUploader(config.Uploader)

	agent := &Agent{
		artGenerator: config.ArtGenerator,
		indexer:      indexer,
		ethClient:    config.EthClient,
		wallet:       wallet,
		nftUploader:  nftUploader,
		tappdClient:  config.TappdClient,
		apiRouter:    nil,
		httpClient:   config.HttpClient,

		rsaPrivateKey: rsaPrivateKey,

		factoryAddress:  config.FactoryAddress,
		pollingInterval: config.PollingInterval,
		apiIpPort:       config.ApiIpPort,
	}

	agent.apiRouter = agent.generateRouter()

	return agent, nil
}

func NewAgentConfigFromSetupResult(setupResult *setup.SetupResult) (*AgentConfig, error) {
	if setupResult == nil {
		return nil, errors.New("setup result is nil")
	}

	ethClient, err := ethclient.Dial(setupResult.EthereumRpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ethereum client: %w", err)
	}

	return &AgentConfig{
		ArtGenerator:   art.NewOpenAiGenerator(setupResult.OpenAiApiKey, setupResult.OpenAiModel),
		Uploader:       filestorage.NewPinataUploader(setupResult.PinataJwtKey),
		EthClient:      ethClient,
		TappdClient:    tappd.NewTappdClient(tappd.WithEndpoint(setupResult.DstackTappdEndpoint)),
		FactoryAddress: setupResult.FactoryAddress,
		HttpClient:     http.DefaultClient,

		PollingInterval: 5 * time.Second,
		PrivateKeySeed:  setupResult.PrivateKeySeed,
		ApiIpPort:       setupResult.ApiIpPort,
	}, nil
}

func (a *Agent) Start(ctx context.Context) error {
	a.StartServer(ctx)

	events := make(chan indexer.PromptSuggestion, 1000)
	a.indexer.IndexEvents(ctx, events)

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}
			go a.processEvent(ctx, event)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (a *Agent) processEvent(ctx context.Context, event indexer.PromptSuggestion) {
	collection, err := contractYayoiCollection.NewContractYayoiCollection(event.Log.Address, a.ethClient)
	if err != nil {
		slog.Error("failed to create collection", "error", err)
		return
	}

	domain, err := collection.Eip712Domain(nil)
	if err != nil {
		slog.Error("failed to get eip712 domain", "error", err)
		return
	}

	systemPromptUri, err := collection.SystemPromptUri(nil)
	if err != nil {
		slog.Error("failed to get system prompt uri", "error", err)
		return
	}

	systemPrompt, err := a.readSystemPromptFromUri(ctx, systemPromptUri)
	if err != nil {
		slog.Error("failed to read system prompt", "error", err)
		return
	}

	artUrl, err := a.artGenerator.GenerateUrl(ctx, event.Prompt, string(systemPrompt))
	if err != nil {
		slog.Error("failed to generate art", "error", err)
		return
	}

	ipfsHash, err := a.nftUploader.UploadUrl(ctx, domain.Name, event.Prompt, artUrl)
	if err != nil {
		slog.Error("failed to upload art", "error", err)
		return
	}

	signature, err := a.wallet.SignMintMessage(event.Sender, ipfsHash, wallet.EIP712Domain{
		Name:              domain.Name,
		Version:           domain.Version,
		ChainId:           domain.ChainId,
		VerifyingContract: domain.VerifyingContract,
	})
	if err != nil {
		slog.Error("failed to sign mint message", "error", err)
		return
	}

	a.mu.Lock()
	_, err = collection.MintGeneratedToken(a.wallet.Auth(), event.Sender, ipfsHash, signature)
	if err != nil {
		slog.Error("failed to mint", "error", err)
		return
	}
	a.mu.Unlock()
}

func (a *Agent) readSystemPromptFromUri(ctx context.Context, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return "", err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decryptedBody, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, a.rsaPrivateKey, body, nil)
	if err != nil {
		slog.Warn("failed to decrypt body", "error", err)
		decryptedBody = body
	}

	return string(decryptedBody), nil
}

func (a *Agent) FactoryAddress() common.Address {
	return a.factoryAddress
}

func (a *Agent) PollingInterval() time.Duration {
	return a.pollingInterval
}

func (a *Agent) ApiIpPort() string {
	return a.apiIpPort
}
