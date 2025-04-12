package ethfetcher

import (
	"context"
	"ethfetcher/internal/ethereumfetcher/process"
	"ethfetcher/internal/ethereumfetcher/service"
	"ethfetcher/internal/ethereumfetcher/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	ctx context.Context,
	pool *pgxpool.Pool,
	srv *echo.Echo,
) {
	ethStore := store.NewStore(pool)
	ethService := service.NewService(ethStore)
	process.RegisterEthHandlers(ctx, srv, ethService)
}
