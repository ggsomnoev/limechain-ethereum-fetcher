package process

import (
	"context"
	"errors"
	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/logger"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

var ErrTxHashesParamIsRequired = errors.New("transactionHashes query parameter is required")

//counterfeiter:generate . Service
type Service interface {
	GetEthTransactions(context.Context, []string) ([]api.Transaction, error)
	GetEthTransactionsByRLP(context.Context, string) ([]api.Transaction, error)
	GetAllEthTransactions(context.Context) ([]api.Transaction, error)
}

//counterfeiter:generate . TokenValidationService
type TokenValidationService interface {
	ValidateToken(ctx context.Context, token string) (bool, string, error)
}

func RegisterEthHandlers(ctx context.Context, srv *echo.Echo, svc Service, tokenValidationSvc TokenValidationService) {
	if srv != nil {
		srv.GET("/lime/eth", handleEthTransactions(ctx, svc, tokenValidationSvc))
		srv.GET("/lime/eth/:rlphex", handleEthTransactionByRLP(ctx, svc, tokenValidationSvc))
		srv.GET("/lime/all", handleAllEthTransactions(ctx, svc, tokenValidationSvc))
		srv.GET("/lime/my", handleMyTransactions(ctx, svc, tokenValidationSvc))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleEthTransactions(ctx context.Context, svc Service, tokenValidationSvc TokenValidationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		transactionHashes := c.QueryParams()["transactionHashes"]
		if len(transactionHashes) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": ErrTxHashesParamIsRequired.Error(),
			})
		}

		transactions, err := svc.GetEthTransactions(ctx, transactionHashes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		_, err = validateTokenIfPresent(ctx, c.Request().Header, tokenValidationSvc)
		if err != nil {
			if err.Error() == "invalid or expired token" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// TODO: store the transactions for the authorized user.

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}

func handleEthTransactionByRLP(ctx context.Context, svc Service, tokenValidationSvc TokenValidationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		rlpHex := c.Param("rlphex")

		transactions, err := svc.GetEthTransactionsByRLP(ctx, rlpHex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		_, err = validateTokenIfPresent(ctx, c.Request().Header, tokenValidationSvc)
		if err != nil {
			if err.Error() == "invalid or expired token" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// TODO: store the transactions for the authorized user.

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}

func handleAllEthTransactions(ctx context.Context, svc Service, tokenValidationSvc TokenValidationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		transactions, err := svc.GetAllEthTransactions(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to fetch all transactions: %v", err),
			})
		}

		_, err = validateTokenIfPresent(ctx, c.Request().Header, tokenValidationSvc)
		if err != nil {
			if err.Error() == "invalid or expired token" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// TODO: store the transactions for the authorized user.

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}

func handleMyTransactions(ctx context.Context, svc Service, tokenValidationSvc TokenValidationService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username, err := validateTokenIfPresent(ctx, c.Request().Header, tokenValidationSvc)
		if err != nil {
			if err.Error() == "invalid or expired token" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		if username == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "authorization token is required",
			})
		}

		// TODO: get the transactions for the authorized user.

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": []api.Transaction{},
		})
	}
}

func extractTokenFromHeader(header http.Header) string {
	return strings.TrimPrefix(header.Get("Authorization"), "Bearer ")
}

func validateTokenIfPresent(ctx context.Context,
	header http.Header,
	tokenValidationSvc TokenValidationService,
) (string, error) {
	token := extractTokenFromHeader(header)
	if token == "" {
		return "", nil
	}

	valid, username, err := tokenValidationSvc.ValidateToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}
	if !valid {
		return "", errors.New("invalid or expired token")
	}

	return username, nil
}
