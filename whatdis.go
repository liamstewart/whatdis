package whatdis

import (
	"fmt"
	rand "math/rand/v2"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/liamstewart/whatdis/internal"
	"go.uber.org/zap"
)

type Config struct {
	Endpoints map[string]Endpoint
}

type Endpoint struct {
	Path       string
	Status     int
	Response   Response
	Methods    []string
	Middleware []Middleware
}

type Middleware struct {
	Middleware string
	Sleep      Sleep
	Fail       Fail
}

type Sleep struct {
	Distribution string
	Uniform      Uniform
	Normal       Normal
}

type Fail struct {
	Distribution string
	Bernoulli    Bernoulli
}

type Uniform struct {
	A int64
	B int64
}

type Normal struct {
	Mean   float64
	StdDev float64
}

type Bernoulli struct {
	P float64
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

func (s *Whatdis) badRequest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "%s", "bad request")
}

func (s *Whatdis) addHandler(endpoint *Endpoint) {
	m := func(w http.ResponseWriter, req *http.Request) {
		if len(endpoint.Methods) > 0 && !slices.Contains(endpoint.Methods, req.Method) {
			s.badRequest(w, req)
			return
		}

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

	var h http.Handler
	h = http.HandlerFunc(m)

	r := rand.New(rand.NewPCG(1, 2))

	for i := 0; i < len(endpoint.Middleware); i++ {
		m := endpoint.Middleware[i]
		if m.Middleware == "sleep" {
			var d internal.RandomVariable[int64]
			if m.Sleep.Distribution == "uniform" {
				d = internal.NewUniform(
					m.Sleep.Uniform.A,
					m.Sleep.Uniform.B,
					r,
				)
			} else if m.Sleep.Distribution == "normal" {
				d = internal.NewNormal(
					m.Sleep.Normal.Mean,
					m.Sleep.Normal.StdDev,
					r,
				)
			} else {
				panic("unsupported distribution")
			}
			h = internal.SleepHandler(h, d)
		} else if m.Middleware == "fail" {
			var d internal.RandomVariable[bool]
			if m.Fail.Distribution == "bernoulli" {
				d = internal.NewBernoulli(m.Fail.Bernoulli.P, r)
			} else {
				panic("unsupported distribution")
			}
			h = internal.FailHandler(h, d)
		}
	}

	s.mux.Handle(fmt.Sprintf("/%s", endpoint.Path), h)
}
