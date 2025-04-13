package auth

import (
	"context"
	"ethfetcher/internal/auth/process"
	"ethfetcher/internal/auth/service"
	"ethfetcher/internal/auth/store"
	"ethfetcher/internal/lifecycle"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	ctx context.Context,
	procSpawnFn lifecycle.ProcessSpawnFunc,
	pool *pgxpool.Pool,
	srv *echo.Echo,
	secret string,
) *service.Service {
	authStore := store.NewStore(pool)
	authService := service.NewService(authStore, secret)

	// Process for removing expired tokens.
	process.Process(procSpawnFn, authService)

	process.RegisterAuthHandlers(ctx, srv, authService)

	return authService
}
