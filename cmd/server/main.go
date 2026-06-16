// Command server runs the aivaldebatten backend: it loads config, starts the
// broadcast hub and the debate engine, and serves the WebSocket stream plus the
// embedded Svelte UI over HTTP.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"aivaldebatten/internal/config"
	"aivaldebatten/internal/engine"
	"aivaldebatten/internal/hub"
	"aivaldebatten/internal/provider"
	"aivaldebatten/internal/server"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	cfgPath := flag.String("config", "config/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("loaded %d bots, %d topics, cycle=%s",
		len(cfg.Bots), len(cfg.Topics), cfg.CycleDuration.Std())

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the broadcast hub.
	h := hub.New()
	go h.Run(ctx)

	// Start the debate engine with the mock provider (swap for an LLM later).
	eng := engine.New(cfg, provider.NewMock(), h)
	go eng.Run(ctx)

	// HTTP server.
	srv := &http.Server{
		Addr:              *addr,
		Handler:           server.Handler(h),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the listener in a goroutine so we can wait for the shutdown signal.
	go func() {
		log.Printf("listening on %s", *addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: %v", err)
		}
	}()

	// Block until a shutdown signal arrives.
	<-ctx.Done()
	log.Printf("shutting down...")

	// Give in-flight requests a short grace period to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
