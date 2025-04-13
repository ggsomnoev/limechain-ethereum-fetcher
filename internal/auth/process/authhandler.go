package process

import (
	"context"
	"errors"
	"ethfetcher/internal/auth/model"
	"ethfetcher/internal/auth/service"
	"ethfetcher/internal/logger"
	"net/http"

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
		var req model.AuthRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		token, err := svc.Authenticate(ctx, req.Username, req.Password)
		if err != nil {
			if errors.Is(err, service.ErrInvalidCredentials) {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid credentials",
				})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, model.AuthResponse{Token: token})
	}
}
