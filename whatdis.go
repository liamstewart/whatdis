package whatdis

import (
	"fmt"
	"net/http"

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

func (s *Whatdis) Handler() http.Handler {
	return s.mux
}

func NewWhatdis(config *Config, logger *zap.SugaredLogger) *Whatdis {
	mux := http.NewServeMux()

	s := &Whatdis{
		logger: logger,
		mux:    mux,
	}

	for _, v := range config.Endpoints {
		mux.HandleFunc(fmt.Sprintf("/%s", v.Path), makeHandler(v.Status, &v.Response))
	}

	return s
}

func makeHandler(status int, resp *Response) func(http.ResponseWriter, *http.Request) {
	h := func(w http.ResponseWriter, req *http.Request) {
		accept := "text/plain"
		accepts := req.Header["Accept"]
		if len(accepts) > 0 {
			accept = accepts[0]
		}

		headers := w.Header()
		if accept == "application/json" {
			headers.Set("Accept", "application/json")
		}
		w.WriteHeader(status)

		if accept == "application/json" {
			fmt.Fprintf(w, "%s", resp.Json)
		} else {
			fmt.Fprintf(w, "%s", resp.Text)
		}
	}
	return h
}
