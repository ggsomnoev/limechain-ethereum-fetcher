package process

import (
	"context"
	"ethfetcher/internal/logger"

	"github.com/labstack/echo/v4"
)

func RegisterAuthHandlers(ctx context.Context, srv *echo.Echo, svc Service) {
	if srv != nil {
		srv.POST("/lime/authenticate", handleAuthentication(ctx, svc))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleAuthentication(ctx context.Context, svc Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}
