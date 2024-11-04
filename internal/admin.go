package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Admin struct {
	logger *zap.SugaredLogger
	stop   context.CancelFunc
	mux    http.Handler
}

func NewAdmin(stop context.CancelFunc, logger *zap.SugaredLogger) *Admin {
	mux := http.NewServeMux()

	s := &Admin{
		logger: logger,
		stop:   stop,
		mux:    mux,
	}

	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/env", s.handleEnv)
	mux.HandleFunc("/ping", s.handlePing)
	mux.HandleFunc("/livez", s.handleLivez)
	mux.HandleFunc("/readyz", s.handleReadyz)
	mux.HandleFunc("/quitquitquit", s.handleQuit)
	mux.HandleFunc("/abortabortabort", s.handleAbort)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/deubg/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/deubg/pprof/profile", pprof.Profile)
	mux.HandleFunc("/deubg/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/deubg/pprof/trace", pprof.Trace)

	mux.Handle("/metrics", promhttp.Handler())

	return s
}

func (s *Admin) Handler() http.Handler {
	return s.mux
}

func (s *Admin) JSONResponse(w http.ResponseWriter, r *http.Request, result interface{}) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("JSON marshal failed", zap.Error(err))
		return
	}
	body = prettify(body)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (s *Admin) handleInfo(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	data["command"] = strings.Join(os.Args, " ")
	data["runtime.NumCPU"] = fmt.Sprintf("%d", runtime.NumCPU())
	data["runtime.Version"] = runtime.Version()
	data["runtime.GOMAXPROCS"] = fmt.Sprintf("%d", runtime.GOMAXPROCS(0))
	data["runtime.GOARCH"] = runtime.GOARCH
	data["runtime.GOOS"] = runtime.GOOS

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			data[fmt.Sprintf("build.%s", setting.Key)] = setting.Value
		}
	}

	s.JSONResponse(w, r, data)
}

func (s *Admin) handleEnv(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, r, os.Environ())
}

func (s *Admin) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "pong")
}

func (s *Admin) handleLivez(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, r, map[string]string{"status": "OK"})
}

func (s *Admin) handleReadyz(w http.ResponseWriter, r *http.Request) {
	s.JSONResponse(w, r, map[string]string{"status": "OK"})
}

func (s *Admin) handleQuit(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("quit received")
	s.stop()
	s.JSONResponse(w, r, map[string]string{"status": "OK"})
}

func (s *Admin) handleAbort(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("abort received")
	os.Exit(255)
}
