package service

import (
	"context"
	"errors"
	"fmt"

	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/logger"
)

//counterfeiter:generate . Store
type Store interface {
	GetAll(context.Context) ([]api.Transaction, error)
	GetByHash(ctx context.Context, hash string) (api.Transaction, error)
	Insert(context.Context, api.Transaction) error
}

//counterfeiter:generate . EthereumClient
type EthereumClient interface {
	FetchTransactionByHash(ctx context.Context, hash string) (api.Transaction, error)
	RlpHexToHashList(rlpHex string) ([]string, error)
}

type Service struct {
	store     Store
	ethClient EthereumClient
}

var ErrTransactionNotFound = errors.New("failed to fetch transactions")

func NewService(store Store, ethClient EthereumClient) *Service {
	return &Service{
		store:     store,
		ethClient: ethClient,
	}
}

func (s *Service) GetEthTransactions(ctx context.Context, transactionHashes []string) ([]api.Transaction, error) {
	var transactions []api.Transaction

	for _, hash := range transactionHashes {
		tx, err := s.store.GetByHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction %s: %w", hash, err)
		}

		if tx.TransactionHash == "" {
			tx, err = s.ethClient.FetchTransactionByHash(ctx, hash)
			if err != nil {
				return nil, err
			}

			if tx.TransactionHash == "" {
				logger.GetLogger().Warn(fmt.Errorf("%w: %s not found on the ethereum node", ErrTransactionNotFound, hash))
				continue
			}

			err = s.store.Insert(ctx, tx)
			if err != nil {
				return nil, fmt.Errorf("failed to insert transaction %s: %v", hash, err)
			}
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (s *Service) GetEthTransactionsByRLP(ctx context.Context, rlpHex string) ([]api.Transaction, error) {
	transactionHashes, err := s.ethClient.RlpHexToHashList(rlpHex)
	if err != nil {
		return nil, err
	}

	return s.GetEthTransactions(ctx, transactionHashes)
}

func (s *Service) GetAllEthTransactions(ctx context.Context) ([]api.Transaction, error) {
	return s.store.GetAll(ctx)
}

func (s *Service) GetAllUserTransactions(context.Context, string) ([]api.Transaction, error) {
	return nil, nil
}
func (s *Service) SetUserTransactions(context.Context, string, []api.Transaction) error {
	return nil
}
