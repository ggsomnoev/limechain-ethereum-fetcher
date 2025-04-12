package process

import (
	"context"
	"ethfetcher/internal/ethereumfetcher/model"
	"ethfetcher/internal/logger"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Service interface {
	GetEthTransactions(ctx context.Context, transactionHashes []string) ([]model.Transaction, error)
	GetEthTransactionsByRLP(ctx context.Context, rlpHex string) ([]model.Transaction, error)
	GetAllEthTransactions(ctx context.Context) ([]model.Transaction, error)
}

func RegisterEthHandlers(ctx context.Context, srv *echo.Echo, svc Service) {
	if srv != nil {
		srv.GET("/lime/eth", handleEthTransactions(ctx, svc))
		srv.GET("/lime/eth/:rlphex", handleEthTransactionByRLP(ctx, svc))
		srv.GET("/lime/all", handleAllEthTransactions(ctx, svc))
	} else {
		logger.GetLogger().Warn("Running routes without a webapi server, did NOT register routes.")
	}
}

func handleEthTransactions(ctx context.Context, svc Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		transactionHashes := c.QueryParams()["transactionHashes"]
		if len(transactionHashes) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "transactionHashes query parameter is required",
			})
		}

		transactions, err := svc.GetEthTransactions(ctx, transactionHashes)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to fetch transactions: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}

func handleEthTransactionByRLP(ctx context.Context, svc Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		rlpHex := c.Param("rlphex")

		transactions, err := svc.GetEthTransactionsByRLP(ctx, rlpHex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to fetch transactions: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}

func handleAllEthTransactions(ctx context.Context, svc Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		transactions, err := svc.GetAllEthTransactions(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("failed to fetch all transactions: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"transactions": transactions,
		})
	}
}
