package store

import (
	"context"
	"fmt"
)

func (s *Store) DeleteTxsByHash(ctx context.Context, hash string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE transaction_hash = $1`, TransactionTable)

	_, err := s.pool.Exec(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("failed to delete transaction from DB: %w", err)
	}

	return nil
}

func (s *Store) DeleteTxsByUsername(ctx context.Context, username string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE username = $1`, UserTransactionsTable)

	_, err := s.pool.Exec(ctx, query, username)
	if err != nil {
		return fmt.Errorf("failed to delete transaction from DB: %w", err)
	}

	return nil
}
