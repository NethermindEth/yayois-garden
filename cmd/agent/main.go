package main

import (
	"context"
	"log/slog"

	"github.com/NethermindEth/yayois-garden/pkg/agent"
	"github.com/NethermindEth/yayois-garden/pkg/agent/setup"
)

func main() {
	ctx := context.Background()

	setupResult, err := setup.Setup(ctx)
	if err != nil {
		slog.Error("failed to setup", "error", err)
		return
	}

	agent, err := agent.NewAgent(ctx, setupResult)
	if err != nil {
		slog.Error("failed to create agent", "error", err)
		return
	}

	agent.Start(ctx)
}
