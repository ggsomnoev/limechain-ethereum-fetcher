package process

import (
	"context"
	"ethfetcher/internal/lifecycle"
	"fmt"
	"time"
)

const pollInterval = 60 * time.Second

//counterfeiter:generate . Service
type Service interface {
	DeleteExpiredTokens(context.Context) error
	Authenticate(context.Context, string, string) (string, error)
}

func Process(
	procSpawnFn lifecycle.ProcessSpawnFunc,
	svc Service,
) {
	procSpawnFn(func(ctx context.Context) error {
		pollChan := time.NewTicker(pollInterval)
		defer pollChan.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-pollChan.C:
				// Go on..
			}

			err := svc.DeleteExpiredTokens(ctx)
			if err != nil {
				return fmt.Errorf("failed to delete expired auth tokens: %w", err)
			}
		}
	}, "Expired token remover")
}
