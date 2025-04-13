package store

import (
	"context"
	"fmt"
)

func (s *Store) DeleteTxByHash(ctx context.Context, hash string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE transaction_hash = $1`, TransactionTable)

	_, err := s.pool.Exec(ctx, query, hash)
	if err != nil {
		return fmt.Errorf("failed to delete transaction from DB: %w", err)
	}

	return nil
}
