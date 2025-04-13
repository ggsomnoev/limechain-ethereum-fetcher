package service

import "context"

//counterfeiter:generate . Store
type Store interface {
}

type Service struct {
	store  Store
	secret string
}

func NewService(store Store, secret string) *Service {
	return &Service{
		store:  store,
		secret: secret,
	}
}

func (s *Service) DeleteExpiredTokens(ctx context.Context) error {
	return nil
}
