package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Service interface {
	Handler() http.Handler
}

func ListenAndServe(h http.Handler, port int, stop context.CancelFunc, logger *zap.SugaredLogger) *http.Server {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h,
	}

	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("finished serving")
		} else {
			logger.Info("failed to serve", zap.Error(err))
			stop()
		}
	}()

	return server
}
