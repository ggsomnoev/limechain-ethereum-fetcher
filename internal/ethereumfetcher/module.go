package ethfetcher

import (
	"context"
	"ethfetcher/internal/ethclient"
	"ethfetcher/internal/ethereumfetcher/process"
	"ethfetcher/internal/ethereumfetcher/service"
	"ethfetcher/internal/ethereumfetcher/store"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

const (
	maxRetries = 5
	retryDelay = 2 * time.Second
)

func Process(
	ctx context.Context,
	pool *pgxpool.Pool,
	srv *echo.Echo,
	ethNodeURL string,
) error {
	ethClient, err := ethclient.NewEthereumClientWithRetry(ctx, ethNodeURL, maxRetries, retryDelay)
	if err != nil {
		return err
	}

	ethStore := store.NewStore(pool)
	ethService := service.NewService(ethStore, ethClient)
	process.RegisterEthHandlers(ctx, srv, ethService)

	return nil
}
