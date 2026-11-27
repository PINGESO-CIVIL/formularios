// Comando api: solo wiring de configuraciÃÆ’Â³n, adaptadores y servidor HTTP.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	httpadapter "github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/api"
	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/config"
	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/store"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api terminÃÆ’Â³ con error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) == 3 && os.Args[1] == "openapi" {
		contrato, err := httpadapter.Contrato()
		if err != nil {
			return err
		}
		return os.WriteFile(os.Args[2], contrato, 0644)
	}
	cfg, err := config.FromOS()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	router := httpadapter.NewRouter(httpadapter.RouterDeps{
		Readiness: []httpadapter.ReadinessCheck{{Name: "postgres", Check: pool.Ping}},
		Metrics:   registry, PublicarDocs: cfg.Env != config.EnvProd,
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api escuchando", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("apagando api")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
