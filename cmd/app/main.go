// Command app runs HexaScore: the HTTP server by default, or a command such
// as migrate, route:list or about (vukem <command> delegates here).
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func main() {
	// Only the server supervises Node; commands (migrate, about, ...) never do.
	if len(os.Args) != 1 || os.Getenv("SSR_SUPERVISE") != "true" {
		os.Exit(vukem.Main(bootstrap.New))
	}

	// Node is not tied to SIGTERM: vukem.Main drains HTTP within
	// SHUTDOWN_TIMEOUT first, and those requests may still render through SSR.
	ctx, stop := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		defer close(done)
		ssr.Supervise(ctx, time.Second, "node", "bootstrap/ssr/ssr.js")
	}()

	code := vukem.Main(bootstrap.New)
	slog.Info("http server stopped, stopping ssr server")
	stop()
	<-done
	slog.Info("ssr server stopped")
	os.Exit(code)
}
