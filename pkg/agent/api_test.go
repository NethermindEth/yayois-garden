package agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/NethermindEth/yayois-garden/pkg/agent"
)

func setupTestAgent(t *testing.T, opts ...func(*agent.AgentConfig)) *agent.Agent {
	mockEthClient, _ := newMockEthClient()

	agentConfig := &agent.AgentConfig{
		ArtGenerator:          &mockArtGenerator{},
		Uploader:              &mockUploader{},
		EthClient:             mockEthClient,
		TappdClient:           &mockTappdClient{},
		FactoryAddress:        common.HexToAddress("0x1234567890123456789012345678901234567890"),
		PollingInterval:       5 * time.Second,
		AccountPrivateKeySeed: agentPrivateKeySeed[:],
		RsaPrivateKey:         rsaPrivateKey,
		ApiIpPort:             "",
	}

	for _, opt := range opts {
		opt(agentConfig)
	}

	agent, err := agent.NewAgent(context.Background(), agentConfig)
	require.NoError(t, err)
	return agent
}

func TestAgentApi_GetRouter(t *testing.T) {
	testAgent := setupTestAgent(t, func(config *agent.AgentConfig) {
		config.TappdClient = &mockTappdClient{
			tdxQuote: func(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error) {
				return &tappd.TdxQuoteResponse{
					Quote: "test-quote",
				}, nil
			},
		}
	})
	router := testAgent.GetRouter()

	t.Run("GET /address", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/address", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testAgent.Address().String(), w.Body.String())
	})

	t.Run("GET /quote", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/quote", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var quote string
		err := json.NewDecoder(w.Body).Decode(&quote)
		assert.NoError(t, err)
		assert.Equal(t, "test-quote", quote)
	})

	t.Run("GET /quote error", func(t *testing.T) {
		testAgent := setupTestAgent(t, func(config *agent.AgentConfig) {
			config.TappdClient = &mockTappdClient{
				tdxQuote: func(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error) {
					return nil, assert.AnError
				},
			}
		})

		router := testAgent.GetRouter()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/quote", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, assert.AnError.Error(), w.Body.String())
	})

	t.Run("GET /pubkey", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/pubkey", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var pubKey map[string]string
		err := json.NewDecoder(w.Body).Decode(&pubKey)
		assert.NoError(t, err)

		assert.Equal(t, rsaPrivateKey.PublicKey.N.String(), pubKey["n"])
		assert.Equal(t, strconv.Itoa(rsaPrivateKey.PublicKey.E), pubKey["e"])
	})
}
