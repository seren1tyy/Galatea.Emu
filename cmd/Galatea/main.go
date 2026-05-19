package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"Galatea.Emu/internal/config"
	"Galatea.Emu/internal/server"
)

func main() {
	// 1. Инициализация логгера
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// 2. Загрузка конфига
	cfg, err := config.Load("configs/server.yaml")
	if err != nil {
		logger.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	// 3. Настройка контекста для graceful shutdown (по Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. Запуск сервера
	srv := server.New(cfg.Server.Host, cfg.Server.Port, logger)

	logger.Info("Starting Galatea Emulator...")
	if err := srv.Start(ctx); err != nil {
		logger.Error("Server stopped with error", "err", err)
	}

	logger.Info("Server shutdown complete")
}
