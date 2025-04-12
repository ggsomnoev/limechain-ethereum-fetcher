package webapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"ethfetcher/internal/lifecycle"

	"github.com/labstack/echo/v4"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
)

func NewServer(ctx context.Context) *echo.Echo {
	e := echo.New()
	e.HidePort = true
	e.HideBanner = true

	return e
}

func Start(procSpawnFn lifecycle.ProcessSpawnFunc, srv *echo.Echo, bindAddress string) {
	startServer(procSpawnFn, srv, bindAddress)
	stopServer(procSpawnFn, srv)
}

func startServer(procSpawnFn lifecycle.ProcessSpawnFunc, e *echo.Echo, bindAddress string) {
	procSpawnFn(func(ctx context.Context) error {
		log.Printf("starting the WebAPI server@%s\n", bindAddress)

		err := e.Start(bindAddress)
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("could not webAPI server: %w", err)
		}

		return nil
	}, "WebAPI Starter")
}

// With graceful shut down.
func stopServer(procSpawnFn lifecycle.ProcessSpawnFunc, e *echo.Echo) {
	procSpawnFn(func(ctx context.Context) error {
		<-ctx.Done()
		log.Println("stopping the WebAPI server due to app exit")

		ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		err := e.Shutdown(ctxGrace)
		if err != nil {
			return fmt.Errorf("failed to shutdown of webAPI server: %w", err)
		}

		return nil
	}, "WebAPI Stopper")
}
