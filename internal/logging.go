package internal

import (
	"net/http"

	"go.uber.org/zap"
)

func LoggingHandler(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	m := func(w http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(w, req)
		logger.Info(
			zap.String("request_method", req.Method),
		)
	}

	return http.HandlerFunc(m)
}
