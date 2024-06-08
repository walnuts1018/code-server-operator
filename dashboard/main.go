package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/walnuts1018/code-server-operator/dashboard/config"
	"github.com/walnuts1018/code-server-operator/dashboard/router"
	"github.com/walnuts1018/code-server-operator/dashboard/router/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to load config: %v", err))
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.RFC3339,
		Level:      cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	handler := handler.NewHandler()

	r, err := router.NewRouter(cfg, *handler)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to create router: %v", err))
		os.Exit(1)
	}

	if err := r.Run(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
		slog.Error(fmt.Sprintf("Failed to start server: %v", err))
		os.Exit(1)
	}
}
