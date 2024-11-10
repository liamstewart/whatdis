package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/honeycombio/otel-config-go/otelconfig"
	"github.com/liamstewart/whatdis"
	"github.com/liamstewart/whatdis/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	undoStdLogger := zap.RedirectStdLog(logger)
	defer undoStdLogger()
	sugar := logger.Sugar()

	otelShutdown, err := otelconfig.ConfigureOpenTelemetry()
	if err != nil {
		sugar.Fatalf("error setting up OTel SDK - %e", err)
	}
	defer otelShutdown()

	ctx, stop := context.WithCancel(context.Background())

	adminService := internal.NewAdmin(stop, sugar)
	wrappedAdminHandler := otelhttp.NewHandler(
		internal.LoggingHandler(adminService.Handler(), sugar, "admin"),
		"admin",
	)
	adminServer := internal.ListenAndServe(wrappedAdminHandler, 8081, stop, sugar)

	f := "example.toml"
	var config whatdis.Config
	_, err = toml.DecodeFile(f, &config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	whatdisService := whatdis.NewWhatdis(&config, sugar)
	wrappedWhatdisHandler := otelhttp.NewHandler(
		internal.LoggingHandler(whatdisService.Handler(), sugar, "server"),
		"server",
	)
	whatdisServer := internal.ListenAndServe(wrappedWhatdisHandler, 8080, stop, sugar)

	internal.SetupSignalHandler(stop)

	<-ctx.Done()

	servers := map[string]*http.Server{
		"admin":   adminServer,
		"whatdis": whatdisServer,
	}
	shutdown := internal.NewShutdown(5*time.Second, sugar)
	shutdown.Graceful(servers)
}
