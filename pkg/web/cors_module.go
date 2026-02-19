package web

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type CORSConfig struct {
	Enabled             bool
	Origins             []string
	Methods             []string
	Headers             []string
	ExposeHeaders       []string
	Credentials         bool
	MaxAgeSeconds       int
	AllowWildcard       bool
	AllowPrivateNetwork bool
}

func CORSModule(cfg CORSConfig) fx.Option {
	if !cfg.Enabled {
		// Tidak pasang middleware jika disabled
		return fx.Options()
	}
	// Build cors.Config
	c := cors.Config{
		AllowOrigins:     cfg.Origins,
		AllowMethods:     defaultIfEmpty(cfg.Methods, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		AllowHeaders:     defaultIfEmpty(cfg.Headers, []string{"Origin", "Authorization", "Content-Type", "X-Requested-With"}),
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.Credentials,
		MaxAge:           time.Duration(max(0, cfg.MaxAgeSeconds)) * time.Second,
	}
	if cfg.AllowWildcard {
		c.AllowWildcard = true
	}
	// Sejak Gin-contrib/cors v1.7.1 ada support Private Network (Chrome)
	// Jika versi tidak mendukung, field ini akan diabaikan.
	if cfg.AllowPrivateNetwork {
		c.AllowPrivateNetwork = true
	}

	return fx.Invoke(func(r *gin.Engine) {
		r.Use(cors.New(c))
	})
}

func defaultIfEmpty(in, def []string) []string {
	if len(in) == 0 {
		return def
	}
	return in
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
