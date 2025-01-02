package agent_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/yayois-garden/pkg/agent"
	"github.com/NethermindEth/yayois-garden/pkg/agent/wallet"
	contractYayoiCollection "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiCollection"
	contractYayoiFactory "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiFactory"
)

var rsaPrivateKey, _ = rsa.GenerateKey(rand.Reader, 2048)

var genesisAccount, _ = crypto.GenerateKey()
var genesisAddress = crypto.PubkeyToAddress(genesisAccount.PublicKey)

var ownerAccount, _ = crypto.GenerateKey()
var ownerAddress = crypto.PubkeyToAddress(ownerAccount.PublicKey)
var ownerAuth, _ = bind.NewKeyedTransactorWithChainID(ownerAccount, big.NewInt(1337))

var userAccount, _ = crypto.GenerateKey()
var userAddress = crypto.PubkeyToAddress(userAccount.PublicKey)
var userAuth, _ = bind.NewKeyedTransactorWithChainID(userAccount, big.NewInt(1337))

var agentPrivateKeySeed = [2048]byte{}
var agentWallet, _ = wallet.NewWallet(agentPrivateKeySeed[:], big.NewInt(1337))
var agentAddress = agentWallet.Address()

type mockArtGenerator struct {
	generateUrl func(ctx context.Context, prompt string, systemPrompt string) (string, error)
}

func (m *mockArtGenerator) GenerateUrl(ctx context.Context, prompt string, systemPrompt string) (string, error) {
	return m.generateUrl(ctx, prompt, systemPrompt)
}

type mockUploader struct {
	uploadUrl  func(ctx context.Context, url string) (string, error)
	uploadJson func(ctx context.Context, json interface{}) (string, error)
}

func (m *mockUploader) UploadUrl(ctx context.Context, url string) (string, error) {
	return m.uploadUrl(ctx, url)
}

func (m *mockUploader) UploadJson(ctx context.Context, json interface{}) (string, error) {
	return m.uploadJson(ctx, json)
}

type mockTappdClient struct {
	tdxQuote             func(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error)
	deriveKeyWithSubject func(ctx context.Context, path string, subject string) (*tappd.DeriveKeyResponse, error)
}

func (m *mockTappdClient) TdxQuote(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error) {
	return m.tdxQuote(ctx, reportData)
}

func (m *mockTappdClient) DeriveKeyWithSubject(ctx context.Context, path string, subject string) (*tappd.DeriveKeyResponse, error) {
	return m.deriveKeyWithSubject(ctx, path, subject)
}

func newMockEthClient() (agent.AgentEthClient, *simulated.Backend, *mockAgentClock) {
	mockBackend := simulated.NewBackend(
		types.GenesisAlloc{
			genesisAddress: {Balance: big.NewInt(1000000000000000000)},
			ownerAddress:   {Balance: big.NewInt(1000000000000000000)},
			userAddress:    {Balance: big.NewInt(1000000000000000000)},
			agentAddress:   {Balance: big.NewInt(1000000000000000000)},
		},
	)
	return mockBackend.Client(), mockBackend, &mockAgentClock{backend: mockBackend}
}

type mockAgentClock struct {
	backend *simulated.Backend
}

func (m *mockAgentClock) Now() time.Time {
	blockNumber, err := m.backend.Client().BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	block, err := m.backend.Client().BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		panic(err)
	}
	return time.Unix(int64(block.Time()), 0)
}

func TestNewAgent(t *testing.T) {
	mockEthClient, _, _ := newMockEthClient()

	tests := []struct {
		name        string
		agentConfig *agent.AgentConfig
		wantErr     bool
	}{
		{
			name: "valid config",
			agentConfig: &agent.AgentConfig{
				ArtGenerator:           &mockArtGenerator{},
				Uploader:               &mockUploader{},
				EthClient:              mockEthClient,
				TappdClient:            &mockTappdClient{},
				FactoryAddress:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
				EventPollingInterval:   5 * time.Second,
				AuctionPollingInterval: 1 * time.Minute,
				AccountPrivateKeySeed:  agentPrivateKeySeed[:],
				RsaPrivateKey:          rsaPrivateKey,
				ApiIpPort:              "",
			},
			wantErr: false,
		},
		{
			name:        "nil config",
			agentConfig: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := agent.NewAgent(context.Background(), tt.agentConfig)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, agent)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, agent)
			}
		})
	}
}

func TestAgent_Start(t *testing.T) {
	mockEthClient, _, _ := newMockEthClient()

	agentConfig := &agent.AgentConfig{
		ArtGenerator:           &mockArtGenerator{},
		Uploader:               &mockUploader{},
		EthClient:              mockEthClient,
		TappdClient:            &mockTappdClient{},
		FactoryAddress:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		EventPollingInterval:   5 * time.Second,
		AuctionPollingInterval: 1 * time.Minute,
		AccountPrivateKeySeed:  agentPrivateKeySeed[:],
		RsaPrivateKey:          rsaPrivateKey,
		ApiIpPort:              "",
	}

	agent, err := agent.NewAgent(context.Background(), agentConfig)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = agent.Start(ctx)
	assert.Error(t, err, context.DeadlineExceeded)
}

func TestAgent_Quote(t *testing.T) {
	var a *agent.Agent
	var err error

	mockEthClient, _, _ := newMockEthClient()

	mockTappdClient := &mockTappdClient{
		tdxQuote: func(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error) {
			writer := bytes.NewBuffer([]byte{})

			binary.Write(writer, binary.BigEndian, a.Address().Bytes())
			binary.Write(writer, binary.BigEndian, a.FactoryAddress().Bytes())

			if !bytes.Equal(reportData, writer.Bytes()) {
				return nil, assert.AnError
			}

			return &tappd.TdxQuoteResponse{
				Quote: "test-quote",
			}, nil
		},
	}

	agentConfig := &agent.AgentConfig{
		ArtGenerator:           &mockArtGenerator{},
		Uploader:               &mockUploader{},
		EthClient:              mockEthClient,
		TappdClient:            mockTappdClient,
		FactoryAddress:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		EventPollingInterval:   5 * time.Second,
		AuctionPollingInterval: 1 * time.Minute,
		AccountPrivateKeySeed:  agentPrivateKeySeed[:],
		RsaPrivateKey:          rsaPrivateKey,
		ApiIpPort:              "",
	}

	a, err = agent.NewAgent(context.Background(), agentConfig)
	require.NoError(t, err)

	quote, err := a.Quote(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "test-quote", quote)
}

func TestAgent_MainFlow(t *testing.T) {
	t.Run("plain text system prompt", func(t *testing.T) {
		mockEthClient, simBackend, simClock := newMockEthClient()

		factoryAddr, tx, factoryInstance, err := contractYayoiFactory.DeployContractYayoiFactory(
			ownerAuth,
			mockEthClient,
			common.HexToAddress("0x0000000000000000000000000000000000000000"),
			big.NewInt(10),
			big.NewInt(1),
			uint64(1),
			ownerAddress,
		)
		require.NoError(t, err)
		simBackend.Commit()

		require.NotEqual(t, factoryAddr, common.Address{}, "Factory address should not be zero")
		require.NotNil(t, factoryInstance, "Factory instance should not be nil")
		require.NotNil(t, tx, "Should have a valid deploy transaction")

		tx2, err := factoryInstance.UpdateAuthorizedSigner(ownerAuth, agentAddress, true)
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx2, "Should have a valid transaction updating the authorized signer")

		tx2Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx2)
		require.NoError(t, err)
		require.NotNil(t, tx2Receipt, "Should have a valid transaction receipt")

		systemPrompt := "test system prompt"
		systemPromptUri := "ipfs://demo"
		userPrompt := "test user prompt"
		artUri := "test-art-uri"
		uploadedArtUri := "test-uploaded-art-uri"
		uploadedJsonUri := "test-uploaded-json-uri"
		collectionName := "test-collection-name"
		collectionSymbol := "TEST"

		mockHttpClient := &http.Client{
			Transport: &mockHttpTransport{
				roundTrip: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == systemPromptUri {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(systemPrompt)),
						}, nil
					}
					return nil, fmt.Errorf("unexpected request to %s", req.URL)
				},
			},
		}

		tx3Params := *ownerAuth
		tx3Params.Value = big.NewInt(10)

		tx3, err := factoryInstance.CreateCollection(&tx3Params, contractYayoiFactory.YayoiFactoryCreateCollectionParams{
			Name:            collectionName,
			Symbol:          collectionSymbol,
			SystemPromptUri: systemPromptUri,
			PaymentToken:    common.Address{},
			MinimumBidPrice: big.NewInt(20),
			AuctionDuration: 3600, // 1 hour auction duration
		})
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx3, "Should have a valid transaction creating a collection")

		tx3Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx3)
		require.NoError(t, err)
		require.NotNil(t, tx3Receipt, "Should have a valid transaction receipt")

		testAgent := setupTestAgent(t, func(config *agent.AgentConfig) {
			config.EthClient = mockEthClient
			config.HttpClient = mockHttpClient
			config.FactoryAddress = factoryAddr
			config.EventPollingInterval = 1 * time.Second
			config.AuctionPollingInterval = 1 * time.Second
			config.ArtGenerator = &mockArtGenerator{
				generateUrl: func(ctx context.Context, prompt string, systemPrompt string) (string, error) {
					require.Equal(t, prompt, userPrompt)
					require.Equal(t, systemPrompt, systemPrompt)
					return artUri, nil
				},
			}
			config.Uploader = &mockUploader{
				uploadUrl: func(ctx context.Context, url string) (string, error) {
					require.Equal(t, url, artUri)
					return uploadedArtUri, nil
				},
				uploadJson: func(ctx context.Context, json interface{}) (string, error) {
					require.Equal(t, json, map[string]string{
						"name":        collectionName,
						"description": userPrompt,
						"image":       uploadedArtUri,
					})
					return uploadedJsonUri, nil
				},
			}
			config.Clock = simClock
		})

		agentCtx, agentCancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			err := testAgent.Start(agentCtx)
			require.Error(t, err, context.DeadlineExceeded)
		}()

		collectionAddr, err := factoryInstance.GetCollectionFromSystemPromptUri(nil, systemPromptUri)
		require.NoError(t, err)
		require.NotEqual(t, collectionAddr, common.Address{})

		collectionInstance, err := contractYayoiCollection.NewContractYayoiCollection(collectionAddr, mockEthClient)
		require.NoError(t, err)
		require.NotNil(t, collectionInstance)

		tx4Params := *userAuth
		tx4Params.Value = big.NewInt(20)

		currentAuctionId, err := collectionInstance.GetCurrentAuctionId(nil)
		require.NoError(t, err)

		tx4, err := collectionInstance.SuggestPrompt(&tx4Params, currentAuctionId, userPrompt, big.NewInt(20))
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx4, "Should have a valid transaction suggesting a prompt")

		tx4Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx4)
		require.NoError(t, err)
		require.NotNil(t, tx4Receipt, "Should have a valid transaction receipt")

		// Finish the auction
		endTime, err := collectionInstance.GetAuctionEndTime(nil, currentAuctionId)
		require.NoError(t, err)

		// Move time forward to end the auction
		simBackend.AdjustTime(time.Duration(endTime.Int64()-simClock.Now().Unix()) * time.Second)
		simBackend.Commit()

		time.After(2 * time.Second)

		<-agentCtx.Done()
		agentCancel()

		simBackend.Commit()

		token0, err := collectionInstance.TokenURI(nil, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, token0, uploadedJsonUri)
	})

	t.Run("encrypted system prompt", func(t *testing.T) {
		mockEthClient, simBackend, simClock := newMockEthClient()

		// Deploy a new factory
		factoryAddr, tx, factoryInstance, err := contractYayoiFactory.DeployContractYayoiFactory(
			ownerAuth,
			mockEthClient,
			common.HexToAddress("0x0000000000000000000000000000000000000000"),
			big.NewInt(10),
			big.NewInt(1),
			uint64(1),
			ownerAddress,
		)
		require.NoError(t, err)
		simBackend.Commit()

		require.NotEqual(t, factoryAddr, common.Address{}, "Factory address should not be zero")
		require.NotNil(t, factoryInstance, "Factory instance should not be nil")
		require.NotNil(t, tx, "Should have a valid deploy transaction")

		// Authorize the agent
		tx2, err := factoryInstance.UpdateAuthorizedSigner(ownerAuth, agentAddress, true)
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx2, "Should have a valid transaction updating the authorized signer")

		tx2Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx2)
		require.NoError(t, err)
		require.NotNil(t, tx2Receipt, "Should have a valid transaction receipt")

		systemPromptDecrypted := "test system prompt (decrypted)"
		systemPromptUri := "ipfs://demo-encrypted"
		userPrompt := "test user prompt"
		artUri := "test-art-uri"
		uploadedArtUri := "test-uploaded-art-uri"
		uploadedJsonUri := "test-uploaded-json-uri"
		collectionName := "test-collection-name-encrypted"
		collectionSymbol := "TEST"

		systemPromptEncrypted, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &rsaPrivateKey.PublicKey, []byte(systemPromptDecrypted), nil)
		require.NoError(t, err)

		mockHttpClient := &http.Client{
			Transport: &mockHttpTransport{
				roundTrip: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == systemPromptUri {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBuffer(systemPromptEncrypted)),
						}, nil
					}
					return nil, fmt.Errorf("unexpected request to %s", req.URL)
				},
			},
		}

		tx3Params := *ownerAuth
		tx3Params.Value = big.NewInt(10)
		tx3, err := factoryInstance.CreateCollection(&tx3Params, contractYayoiFactory.YayoiFactoryCreateCollectionParams{
			Name:            collectionName,
			Symbol:          collectionSymbol,
			SystemPromptUri: systemPromptUri,
			PaymentToken:    common.Address{},
			MinimumBidPrice: big.NewInt(20),
			AuctionDuration: 3600, // 1 hour auction duration
		})
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx3, "Should have a valid transaction creating a collection")

		tx3Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx3)
		require.NoError(t, err)
		require.NotNil(t, tx3Receipt, "Should have a valid transaction receipt")

		testAgent := setupTestAgent(t, func(config *agent.AgentConfig) {
			config.EthClient = mockEthClient
			config.HttpClient = mockHttpClient
			config.FactoryAddress = factoryAddr
			config.EventPollingInterval = 1 * time.Second
			config.AuctionPollingInterval = 1 * time.Second
			config.ArtGenerator = &mockArtGenerator{
				generateUrl: func(ctx context.Context, prompt string, systemPrompt string) (string, error) {
					require.Equal(t, prompt, userPrompt)
					require.Equal(t, systemPromptDecrypted, systemPrompt)
					return artUri, nil
				},
			}
			config.Uploader = &mockUploader{
				uploadUrl: func(ctx context.Context, url string) (string, error) {
					require.Equal(t, url, artUri)
					return uploadedArtUri, nil
				},
				uploadJson: func(ctx context.Context, json interface{}) (string, error) {
					require.Equal(t, json, map[string]string{
						"name":        collectionName,
						"description": userPrompt,
						"image":       uploadedArtUri,
					})
					return uploadedJsonUri, nil
				},
			}
			config.Clock = simClock
		})

		agentCtx, agentCancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			err := testAgent.Start(agentCtx)
			require.Error(t, err, context.DeadlineExceeded)
		}()

		collectionAddr, err := factoryInstance.GetCollectionFromSystemPromptUri(nil, systemPromptUri)
		require.NoError(t, err)
		require.NotEqual(t, collectionAddr, common.Address{})

		collectionInstance, err := contractYayoiCollection.NewContractYayoiCollection(collectionAddr, mockEthClient)
		require.NoError(t, err)
		require.NotNil(t, collectionInstance)

		tx4Params := *userAuth
		tx4Params.Value = big.NewInt(20)

		currentAuctionId, err := collectionInstance.GetCurrentAuctionId(nil)
		require.NoError(t, err)

		tx4, err := collectionInstance.SuggestPrompt(&tx4Params, currentAuctionId, userPrompt, big.NewInt(20))
		require.NoError(t, err)
		simBackend.Commit()
		require.NotNil(t, tx4, "Should have a valid transaction suggesting a prompt")

		tx4Receipt, err := bind.WaitMined(context.Background(), simBackend.Client(), tx4)
		require.NoError(t, err)
		require.NotNil(t, tx4Receipt, "Should have a valid transaction receipt")

		// Finish the auction
		endTime, err := collectionInstance.GetAuctionEndTime(nil, currentAuctionId)
		require.NoError(t, err)

		// Move time forward to end the auction
		simBackend.AdjustTime(time.Duration(endTime.Int64()-simClock.Now().Unix()+1) * time.Second)
		simBackend.Commit()

		<-time.After(2 * time.Second)

		<-agentCtx.Done()
		agentCancel()
		simBackend.Commit()

		token0, err := collectionInstance.TokenURI(nil, big.NewInt(0))
		require.NoError(t, err)
		require.Equal(t, token0, uploadedJsonUri)
	})
}

type mockHttpTransport struct {
	roundTrip func(*http.Request) (*http.Response, error)
}

func (m *mockHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTrip(req)
}
