// Command app runs HexaScore: the HTTP server by default, or a command such
// as migrate, route:list or about (vukem <command> delegates here).
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	vukem "github.com/AlexGalhardo/vukem"

	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func main() {
	// Only the server supervises Node; commands (migrate, about, ...) never do.
	if len(os.Args) == 1 && os.Getenv("SSR_SUPERVISE") == "true" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		go ssr.Supervise(ctx, time.Second, "node", "bootstrap/ssr/ssr.js")
	}

	os.Exit(vukem.Main(bootstrap.New))
}
