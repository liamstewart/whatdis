package internal

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Shutdown struct {
	logger  *zap.SugaredLogger
	timeout time.Duration
}

func NewShutdown(timeout time.Duration, logger *zap.SugaredLogger) *Shutdown {
	return &Shutdown{
		logger:  logger,
		timeout: timeout,
	}
}

func (s *Shutdown) Graceful(servers map[string]*http.Server) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	wg := sync.WaitGroup{}

	for what, server := range servers {
		wg.Add(1)
		go s.shutdown(what, s.shutdownHttp(server), ctx, &wg)
	}

	wg.Wait()
	s.logger.Info("shutdown finished")
}

func (s *Shutdown) shutdown(what string, fp func(context.Context), ctx context.Context, wg *sync.WaitGroup) {
	done := make(chan struct{})
	s.logger.Info(fmt.Sprintf("shutdown %s: shutting down", what))

	go func() {
		fp(ctx)
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info(fmt.Sprintf("shutdown %s: finished", what))
	case <-ctx.Done():
		s.logger.Info(fmt.Sprintf("shutdown %s: timeout", what))
	}
	wg.Done()
}

func (s *Shutdown) shutdownHttp(server *http.Server) func(context.Context) {
	return func(ctx context.Context) {
		if err := server.Shutdown(ctx); err != nil {
			s.logger.Warn("http server graceful shutdown failed", zap.Error(err))
		}
	}
}
