package agent

import (
	"context"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"
)

type TappdClient interface {
	DeriveKeyWithSubject(ctx context.Context, path string, subject string) (*tappd.DeriveKeyResponse, error)
	TdxQuote(ctx context.Context, reportData []byte) (*tappd.TdxQuoteResponse, error)
}
