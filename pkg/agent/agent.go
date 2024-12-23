package agent

import (
	"context"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/NethermindEth/yayois-garden/pkg/agent/art"
	"github.com/NethermindEth/yayois-garden/pkg/agent/filestorage"
	"github.com/NethermindEth/yayois-garden/pkg/agent/indexer"
	"github.com/NethermindEth/yayois-garden/pkg/agent/nft"
	"github.com/NethermindEth/yayois-garden/pkg/agent/setup"
	"github.com/NethermindEth/yayois-garden/pkg/agent/wallet"
	contractYayoiCollection "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiCollection"
)

type Agent struct {
	setupResult *setup.SetupResult

	mu sync.Mutex

	artGenerator art.ArtGenerator
	indexer      *indexer.Indexer
	ethClient    *ethclient.Client
	wallet       *wallet.Wallet
	nftUploader  *nft.NftUploader
	tappdClient  *tappd.TappdClient
}

func NewAgent(ctx context.Context, setupResult *setup.SetupResult) (*Agent, error) {
	artGenerator := art.NewOpenAiGenerator(setupResult.OpenAiApiKey, setupResult.OpenAiModel)

	indexer, err := indexer.NewIndexer(indexer.IndexerOptions{
		RpcUrl:          setupResult.EthereumRpcUrl,
		ContractAddress: setupResult.FactoryAddress,
		PollingInterval: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(setupResult.EthereumRpcUrl)
	if err != nil {
		return nil, err
	}

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	wallet, err := wallet.NewWallet(setupResult.PrivateKeySeed, chainID)
	if err != nil {
		return nil, err
	}

	nftUploader := nft.NewNftUploader(
		filestorage.NewPinataUploader(setupResult.PinataJwtKey),
	)

	tappdClient := tappd.NewTappdClient(
		tappd.WithEndpoint(setupResult.DstackTappdEndpoint),
	)

	return &Agent{
		setupResult: setupResult,

		artGenerator: artGenerator,
		indexer:      indexer,
		ethClient:    ethClient,
		wallet:       wallet,
		nftUploader:  nftUploader,
		tappdClient:  tappdClient,
	}, nil
}

func (a *Agent) Start(ctx context.Context) error {
	events := make(chan indexer.PromptSuggestion, 1000)

	a.indexer.IndexEvents(ctx, events)

	for event := range events {
		go a.processEvent(ctx, event)
	}

	return nil
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

	systemPrompt, err := a.readFromUri(ctx, systemPromptUri)
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

	signature, err := a.wallet.SignMintMessage(event.Sender, ipfsHash, apitypes.TypedDataDomain{
		Name:              domain.Name,
		Version:           domain.Version,
		ChainId:           (*math.HexOrDecimal256)(domain.ChainId),
		VerifyingContract: domain.VerifyingContract.String(),
		Salt:              "0x" + hex.EncodeToString(domain.Salt[:]),
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

func (a *Agent) readFromUri(ctx context.Context, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", uri, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (a *Agent) quote(ctx context.Context) (string, error) {
	reportDataBytes, err := generateReportDataBytes(a.wallet.Address(), a.setupResult.FactoryAddress)
	if err != nil {
		return "", err
	}

	quote, err := a.tappdClient.TdxQuote(ctx, reportDataBytes)
	if err != nil {
		return "", err
	}

	return quote.Quote, nil
}
