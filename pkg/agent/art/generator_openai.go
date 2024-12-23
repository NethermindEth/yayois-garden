package art

import (
	"context"
	"fmt"

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

func (g *OpenAiGenerator) GenerateUrl(ctx context.Context, systemPrompt string, prompt string) (string, error) {
	req := openai.ImageRequest{
		Prompt:         generatePrompt(systemPrompt, prompt),
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
		Model:          g.model,
	}

	resp, err := g.client.CreateImage(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Data) == 0 {
		return "", fmt.Errorf("no image data returned")
	}

	return resp.Data[0].URL, nil
}

func generatePrompt(systemPrompt string, prompt string) string {
	return fmt.Sprintf("%s\n\n%s", systemPrompt, prompt)
}
