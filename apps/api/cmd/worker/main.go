// Comando worker: procesos en segundo plano (outbox relay; luego recurrencias).
// En fase 0 solo arranca y espera la señal de término.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/PINGESO-CIVIL/ms-formularios-pingeso/apps/api/internal/config"
)

func main() {
	cfg, err := config.FromOS()
	if err != nil {
		slog.Error("configuración inválida", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("worker iniciado", "env", cfg.Env)
	<-ctx.Done()
	logger.Info("worker detenido")
}
