package art

import "context"

type ArtGenerator interface {
	Generate(ctx context.Context, systemPrompt string, prompt string) ([]byte, error)
}
