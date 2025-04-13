package service

import (
	"context"
	"errors"
	"time"

	"ethfetcher/internal/auth/model"

	"github.com/golang-jwt/jwt/v5"
)

const ttl = 24 * time.Hour

var ErrInvalidCredentials = errors.New("invalid credentials")

//counterfeiter:generate . Store
type Store interface {
	ValidateUser(context.Context, string, string) (bool, error)
	StoreToken(context.Context, model.Token) error
	IsTokenValid(context.Context, string) (bool, string, error)
	DeleteExpiredTokens(context.Context) error
}

type Service struct {
	store  Store
	secret []byte
}

func NewService(store Store, secret string) *Service {
	return &Service{
		store:  store,
		secret: []byte(secret),
	}
}

func (s *Service) DeleteExpiredTokens(ctx context.Context) error {
	return s.store.DeleteExpiredTokens(ctx)
}

func (s *Service) Authenticate(ctx context.Context, username string, password string) (string, error) {
	valid, err := s.store.ValidateUser(ctx, username, password)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", ErrInvalidCredentials
	}

	expiresAt := time.Now().Add(ttl)
	claims := jwt.MapClaims{
		"sub": username,
		"exp": expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return "", err
	}

	err = s.store.StoreToken(ctx, model.Token{
		Value:     signedToken,
		Username:  username,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s *Service) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	return s.store.IsTokenValid(ctx, token)
}
