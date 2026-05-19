package main

import (
"context"
"log"
"os"
"os/signal"
"syscall"

"Galatea.Emu/internal/config"
"Galatea.Emu/internal/server"
)

func main() {
// 1. Инициализация логгера
logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

// 2. Загрузка конфига
cfg, err := config.Load("configs/server.yaml")
if err != nil {
logger.Printf("Failed to load config: %v", err)
os.Exit(1)
}

// 3. Настройка контекста для graceful shutdown (по Ctrl+C)
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()

// 4. Запуск сервера
srv := server.New(cfg.Server.Host, cfg.Server.Port, logger)

logger.Println("Starting Galatea Emulator...")
if err := srv.Start(ctx); err != nil {
logger.Printf("Server stopped with error: %v", err)
}

logger.Println("Server shutdown complete")
}
