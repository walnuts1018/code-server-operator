package config

import (
	"flag"
	"log/slog"
	"strings"

	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	ServerPort string `env:"SERVER_PORT" envDefault:"8080"`

	LogLevelString string     `env:"LOG_LEVEL" envDefault:"info"`
	LogLevel       slog.Level // Parse from logLevelString
}

func Load() (Config, error) {
	serverport := flag.String("port", "8080", "server port")
	flag.Parse()

	cfg := Config{}
	if serverport != nil {
		cfg.ServerPort = *serverport
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	cfg.LogLevel = parseLogLevel(cfg.LogLevelString)

	return cfg, nil
}

func parseLogLevel(str string) slog.Level {
	switch strings.ToLower(str) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn("Invalid log level, use default level: info")
		return slog.LevelInfo
	}
}
