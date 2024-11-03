package whatdis

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	Endpoints map[string]Endpoint
}

type Endpoint struct {
	Path     string
	Status   int
	Response Response
}

type Response struct {
	Text string
	Json string
}

type Whatdis struct {
	logger *zap.SugaredLogger
	mux    *http.ServeMux
}

func NewWhatdis(config *Config, logger *zap.SugaredLogger) *Whatdis {
	mux := http.NewServeMux()

	s := &Whatdis{
		logger: logger,
		mux:    mux,
	}

	for _, endpoint := range config.Endpoints {
		s.addHandler(&endpoint)
	}

	return s
}

func (s *Whatdis) Handler() http.Handler {
	return s.mux
}

func (s *Whatdis) addHandler(endpoint *Endpoint) {
	h := func(w http.ResponseWriter, req *http.Request) {
		accept := "text/plain"
		accepts := req.Header["Accept"]
		if len(accepts) > 0 {
			accept = accepts[0]
		}

		sleep := "0"
		sleeps := req.Header["X-Whatdis-Sleep"]
		if len(sleeps) > 0 {
			sleep = sleeps[0]
		}
		v, err := strconv.ParseInt(sleep, 10, 32)
		if err != nil {
			s.logger.Error("failed to parse", zap.Error(err))
			v = 0
		}
		if v > 0 {
			time.Sleep(time.Duration(v) * time.Millisecond)
		}

		headers := w.Header()
		if accept == "application/json" {
			headers.Set("Accept", "application/json")
		}
		w.WriteHeader(endpoint.Status)

		if accept == "application/json" {
			fmt.Fprintf(w, "%s", endpoint.Response.Json)
		} else {
			fmt.Fprintf(w, "%s", endpoint.Response.Text)
		}
	}

	s.mux.HandleFunc(fmt.Sprintf("/%s", endpoint.Path), h)
}
