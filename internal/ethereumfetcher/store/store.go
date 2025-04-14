package store

import (
	"context"
	"encoding/json"
	"errors"
	"ethfetcher/internal/ethereumfetcher/api"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TransactionTable = "transactions"

var ErrDuplicateTransaction = errors.New("transaction already exists")

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetAll(ctx context.Context) ([]api.Transaction, error) {
	var transactions []api.Transaction
	query := fmt.Sprintf("SELECT transaction_hash, transaction_data FROM %s", TransactionTable)

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions from DB: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tx api.Transaction
		var transactionData []byte

		if err := rows.Scan(&tx.TransactionHash, &transactionData); err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		if err := json.Unmarshal(transactionData, &tx); err != nil {
			return nil, fmt.Errorf("failed to unmarshal transaction data: %w", err)
		}

		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return transactions, nil
}

func (s *Store) GetByHash(ctx context.Context, hash string) (api.Transaction, error) {
	var tx api.Transaction
	var transactionData []byte

	query := fmt.Sprintf(`SELECT transaction_hash, transaction_data
		FROM %s WHERE transaction_hash = $1`, TransactionTable)

	err := s.pool.QueryRow(ctx, query, hash).Scan(&tx.TransactionHash, &transactionData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.Transaction{}, nil
		}
		return api.Transaction{}, fmt.Errorf("failed to fetch transaction from DB: %w", err)
	}

	if err := json.Unmarshal(transactionData, &tx); err != nil {
		return api.Transaction{}, fmt.Errorf("failed to unmarshal transaction data: %w", err)
	}

	return tx, nil
}

func (s *Store) Insert(ctx context.Context, tx api.Transaction) error {
	transactionData, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %w", err)
	}

	insertQuery := fmt.Sprintf(`INSERT INTO %s (transaction_hash, transaction_data) 
		VALUES ($1, $2)`, TransactionTable)

	_, err = s.pool.Exec(ctx, insertQuery, tx.TransactionHash, transactionData)
	if err != nil {
		var pgErr *pgconn.PgError
		// unique_violation error code
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrDuplicateTransaction
		}
		return fmt.Errorf("failed to insert transaction into DB: %w", err)
	}

	return nil
}

func (s *Store) GetAllByUser(ctx context.Context, username string) ([]api.Transaction, error) {
	return nil, nil
}

func (s *Store) InsertUserTransactions(ctx context.Context, username string, transaction []api.Transaction) error {
	return nil
}
