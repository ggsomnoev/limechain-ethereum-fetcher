package eth

import (
	"context"
	"ethfetcher/internal/logger"
	"fmt"
	"time"

	gethclient "github.com/ethereum/go-ethereum/ethclient"
)

type DialConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
}

func InitClientWithRetry(ctx context.Context, ethNodeURL string, cfg DialConfig) (*gethclient.Client, error) {
	var client *gethclient.Client
	var err error

	for i := 0; i < cfg.MaxRetries; i++ {
		client, err = gethclient.DialContext(ctx, ethNodeURL)
		if err == nil {
			return client, nil
		}

		logger.GetLogger().Info(fmt.Sprintf("Retrying ethereum node connection...(%d)", i))

		backoff := cfg.BaseDelay * (1 << i)
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during Ethereum dial: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
}
