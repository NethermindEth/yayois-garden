package agent

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func (a *Agent) generateRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/address", func(c *gin.Context) {
		c.String(http.StatusOK, a.Address().String())
	})

	router.GET("/pubkey", func(c *gin.Context) {
		c.JSON(http.StatusOK, a.rsaPrivateKey.PublicKey)
	})

	router.GET("/quote", func(c *gin.Context) {
		quote, err := a.Quote(c.Request.Context())
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, quote)
	})

	return router
}

func (a *Agent) GetRouter() *gin.Engine {
	return a.apiRouter
}

func (a *Agent) Quote(ctx context.Context) (string, error) {
	reportDataBytes, err := generateReportDataBytes(a.wallet.Address(), a.factoryAddress)
	if err != nil {
		return "", err
	}

	quote, err := a.tappdClient.TdxQuote(ctx, reportDataBytes)
	if err != nil {
		return "", err
	}

	return quote.Quote, nil
}

func (a *Agent) Address() common.Address {
	return a.wallet.Address()
}

func (a *Agent) StartServer(ctx context.Context) error {
	slog.Info("starting server", "address", a.Address().String(), "port", a.apiIpPort)

	if a.apiIpPort == "" {
		slog.Info("api ip port is empty, skipping server")
		return nil
	}

	server := &http.Server{
		Addr:    a.apiIpPort,
		Handler: a.apiRouter,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
	}()

	return nil
}
