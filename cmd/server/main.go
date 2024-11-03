package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/liamstewart/whatdis"
	"github.com/liamstewart/whatdis/internal"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	undoStdLogger := zap.RedirectStdLog(logger)
	defer undoStdLogger()
	sugar := logger.Sugar()

	ctx, stop := context.WithCancel(context.Background())

	adminService := internal.NewAdmin(stop, sugar)
	adminServer := internal.ListenAndServe(adminService, 8081, stop, sugar)

	f := "example.toml"
	var config whatdis.Config
	_, err := toml.DecodeFile(f, &config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	whatdisService := whatdis.NewWhatdis(&config, sugar)
	whatdisServer := internal.ListenAndServe(whatdisService, 8080, stop, sugar)

	internal.SetupSignalHandler(stop)

	<-ctx.Done()

	servers := map[string]*http.Server{
		"admin":   adminServer,
		"whatdis": whatdisServer,
	}
	shutdown := internal.NewShutdown(5*time.Second, sugar)
	shutdown.Graceful(servers)
}
