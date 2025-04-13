package store

import (
	"context"

	"ethfetcher/internal/auth/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ValidateUser(ctx context.Context, username string, password string) (bool, error) {
	return false, nil
}
func (s *Store) StoreToken(ctx context.Context, token model.Token) error {
	return nil
}
func (s *Store) IsTokenValid(ctx context.Context, tokenStr string) (bool, error) {
	return false, nil
}

func (s *Store) DeleteExpiredTokens(context.Context) error {
	return nil
}
