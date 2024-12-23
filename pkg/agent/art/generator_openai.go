package art

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type OpenAiGenerator struct {
	apiKey string
	model  string
	client *openai.Client
}

var _ ArtGenerator = (*OpenAiGenerator)(nil)

func NewOpenAiGenerator(apiKey string, model string) *OpenAiGenerator {
	client := openai.NewClient(apiKey)
	return &OpenAiGenerator{
		apiKey: apiKey,
		model:  model,
		client: client,
	}
}

func (g *OpenAiGenerator) Generate(ctx context.Context, systemPrompt string, prompt string) ([]byte, error) {
	req := openai.ImageRequest{
		Prompt:         generatePrompt(systemPrompt, prompt),
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
		Model:          g.model,
	}

	resp, err := g.client.CreateImage(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no image data returned")
	}

	imageURL := resp.Data[0].URL

	imageBytes, err := fetchImage(ctx, imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %v", err)
	}

	return imageBytes, nil
}

func generatePrompt(systemPrompt string, prompt string) string {
	return fmt.Sprintf("%s\n\n%s", systemPrompt, prompt)
}

func fetchImage(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
