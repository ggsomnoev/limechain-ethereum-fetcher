package store

import (
	"context"
	"errors"
	"ethfetcher/internal/auth/model"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const TokenTable = "auth_tokens"
const UserTable = "users"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Note: for the sake of simplicity we will work with passwords in plain text.
func (s *Store) ValidateUser(ctx context.Context, username, password string) (bool, error) {
	query := fmt.Sprintf(`SELECT 1 FROM %s WHERE username = $1 AND password = $2`, UserTable)
	var exists int
	err := s.pool.QueryRow(ctx, query, username, password).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to validate user: %w", err)
	}
	return true, nil
}

func (s *Store) StoreToken(ctx context.Context, token model.Token) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (token, username, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (username) 
		DO UPDATE 
			SET token = $1, expires_at = $3;
	`, TokenTable)

	_, err := s.pool.Exec(ctx, query, token.Value, token.Username, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	return nil
}

func (s *Store) IsTokenValid(ctx context.Context, token string) (bool, string, error) {
	query := fmt.Sprintf(`SELECT username FROM %s WHERE token = $1 AND expires_at > NOW()`, TokenTable)
	var username string
	err := s.pool.QueryRow(ctx, query, token).Scan(&username)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to check token validity: %w", err)
	}

	return true, username, nil
}

func (s *Store) DeleteExpiredTokens(ctx context.Context) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE expires_at <= NOW()`, TokenTable)
	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	return nil
}
