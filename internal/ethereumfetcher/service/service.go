package service

import (
	"context"

	"ethfetcher/internal/ethereumfetcher/model"
	"ethfetcher/internal/logger"
)

type Store interface {
}

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (*Service) GetEthTransactions(ctx context.Context, transactionHashes []string) ([]model.Transaction, error) {
	logger.GetLogger().Info("fetching eth transactions...")
	return nil, nil
}

func (*Service) GetEthTransactionsByRLP(ctx context.Context, rlpHex string) ([]model.Transaction, error) {
	logger.GetLogger().Info("fetching eth transactions by rlpHex...")
	return nil, nil
}
func (*Service) GetAllEthTransactions(ctx context.Context) ([]model.Transaction, error) {
	logger.GetLogger().Info("fetching all eth transactions...")
	return nil, nil
}
