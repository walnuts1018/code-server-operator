package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"github.com/walnuts1018/code-server-operator/dashboard/config"
	"github.com/walnuts1018/code-server-operator/dashboard/router/handler"
)

func NewRouter(cfg config.Config, handler handler.Handler) (*gin.Engine, error) {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(sloggin.New(slog.Default()))

	if cfg.LogLevel != slog.LevelDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	r.GET(("/"), handler.Home)

	r.GET(("/healthz"), handler.Health)

	admin := r.Group("/admin")
	{
		_ = admin
	}

	return r, nil
}
