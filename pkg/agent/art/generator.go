package art

import "context"

type ArtGenerator interface {
	GenerateUrl(ctx context.Context, systemPrompt string, prompt string) (string, error)
}
