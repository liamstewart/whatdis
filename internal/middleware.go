package internal

import (
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"go.uber.org/zap"
)

func SleepHandler(h http.Handler, d Distribution) http.Handler {
	m := func(w http.ResponseWriter, req *http.Request) {
		v := max(0, d.Sample())
		time.Sleep(time.Duration(v) * time.Millisecond)
		h.ServeHTTP(w, req)
	}

	return http.HandlerFunc(m)
}

func RecoveryHandler(h http.Handler) http.Handler {
	m := func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				// TODO: log
			}
		}()

		h.ServeHTTP(w, req)
	}

	return http.HandlerFunc(m)
}

// TODO: pass additional static fields?
// TODO: force omit caller, other fields?
func LoggingHandler(h http.Handler, logger *zap.SugaredLogger, operation string) http.Handler {
	wrapped := func(w http.ResponseWriter, req *http.Request) {
		m := httpsnoop.CaptureMetrics(h, w, req)
		logger.Infow(
			"request",
			"name", operation,
			"request_url", req.URL,
			"protocol", req.Proto,
			"request_method", req.Method,
			"status_code", m.Code,
			"response_size_bytes", m.Written,
			"response_time_secs", m.Duration,
			"user_id", "-",
			"user_name", "-",
		)
	}

	return http.HandlerFunc(wrapped)
}
