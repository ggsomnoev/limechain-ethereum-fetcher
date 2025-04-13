package main

import (
	"fmt"
	"time"

	"ethfetcher/internal/eth"
	ethfetcher "ethfetcher/internal/ethereumfetcher"
	"ethfetcher/internal/lifecycle"
	"ethfetcher/internal/logger"
	"ethfetcher/internal/pg"
	"ethfetcher/internal/webapi"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	DBConnectionURL   string        `env:"DB_CONNECTION_URL" envDefault:"postgres://ethuser:ethpass@ethfetcherdb:5432/postgres"`
	DBMaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" envDefault:"30m"`
	DBMaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME" envDefault:"5m"`
	DBHealthCheck     time.Duration `env:"DB_HEALTH_CHECK_PERIOD" envDefault:"1m"`
	DBMinConns        int32         `env:"DB_MIN_CONNS" envDefault:"1"`
	DBMaxConns        int32         `env:"DB_MAX_CONNS" envDefault:"2"`

	EthNodeURL    string        `env:"ETH_NODE_URL" envDefault:"https://ethereum-sepolia-rpc.publicnode.com"`
	EthMaxRetries int           `env:"ETH_MAX_RETRIES" envDefault:"5"`
	EthRetryDelay time.Duration `env:"ETH_RETRY_DELAY" envDefault:"2s"`

	APIPort string `env:"API_PORT" envDefault:"8080"`
}

func main() {
	appController := lifecycle.NewController()
	appCtx, procSpawnFn := appController.Start()

	log := logger.GetLogger()

	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(fmt.Errorf("failed reading configuration, exiting - %w", err))
	}

	dbCfg := pg.PoolConfig{
		MinConns:          cfg.DBMinConns,
		MaxConns:          cfg.DBMaxConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheck,
	}

	pool, err := pg.InitPool(appCtx, cfg.DBConnectionURL, dbCfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed initializing db connection pool, exiting - %w", err))
	}

	defer pool.Close()

	srv := webapi.NewServer(appCtx)

	ethCfg := eth.DialConfig{
		MaxRetries: cfg.EthMaxRetries,
		BaseDelay:  cfg.EthRetryDelay,
	}

	ethClient, err := eth.InitClientWithRetry(appCtx, cfg.EthNodeURL, ethCfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed initializing ethereum client, exiting - %w", err))
	}

	ethfetcher.Process(appCtx, pool, srv, ethClient)

	webapi.Start(procSpawnFn, srv, cfg.APIPort)

	appController.Wait()
}
