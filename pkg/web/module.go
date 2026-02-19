package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Config struct{ Port int }
type Router = *gin.Engine

type webModule struct {
	Cfg Config
}

func Module(port int, mounts ...func(r *gin.Engine)) fx.Option {
	m := &webModule{Cfg: Config{Port: port}}
	return fx.Options(
		fx.Provide(func() Router {
			r := gin.New()
			r.Use(gin.Recovery(), gin.Logger())
			for _, mount := range mounts {
				mount(r)
			}
			return r
		}),
		fx.Invoke(m.startHTTP),
	)
}

func (m *webModule) startHTTP(lc fx.Lifecycle, r Router) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", m.Cfg.Port),
		Handler: r,
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error { go srv.ListenAndServe(); return nil },
		OnStop:  func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
}
