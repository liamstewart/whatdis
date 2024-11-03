package internal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT}
var ensureOneSignalHandler = make(chan struct{})

// Set up signal handling for SIGTERM and SIGINT. The first handled
// signal notifies which Signal should be awaited on elsewhere in the
// program to start shutting things down. A second signal terminates
// the program immediately.
func SetupSignalHandler(stop context.CancelFunc) {
	close(ensureOneSignalHandler) // panics when called twice

	c := make(chan os.Signal, 2)
	signal.Notify(c, shutdownSignals...)
	go func() {
		<-c
		stop()
		<-c
		os.Exit(1)
	}()
}
