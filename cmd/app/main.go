package main

import (
	"github.com/gin-gonic/gin"
	"github.com/teten-nugraha/goboot/pkg/logging"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/teten-nugraha/goboot/internal/app"
	"github.com/teten-nugraha/goboot/internal/bootstrap"
	"github.com/teten-nugraha/goboot/internal/config"
	"github.com/teten-nugraha/goboot/pkg/actuator"
	"github.com/teten-nugraha/goboot/pkg/data"
	"github.com/teten-nugraha/goboot/pkg/metrics"
	"github.com/teten-nugraha/goboot/pkg/security"
	"github.com/teten-nugraha/goboot/pkg/web"
	"go.uber.org/zap"
)

func main() {
	profile := config.ActiveProfile()
	cfg, err := config.Load(profile)
	if err != nil {
		panic(err)
	}

	fx.New(
		fx.Supply(*cfg),

		// ✅ Provide zap logger dengan opsi dari config dan fallback aman
		fx.Provide(func() (*zap.Logger, error) {
			logger, logPath, err := logging.NewFileLoggerWithOptions(logging.FileLoggerOptions{
				Dir:        cfg.Logging.Dir,
				Filename:   cfg.Logging.Filename,
				Level:      cfg.Logging.Level,
				MaxSizeMB:  cfg.Logging.MaxSizeMB,
				MaxBackups: cfg.Logging.MaxBackups,
				MaxAgeDays: cfg.Logging.MaxAgeDays,
				Compress:   cfg.Logging.Compress,
				AppName:    cfg.App.Name,
			})
			if err != nil {
				return nil, err
			}
			// Set sebagai global & informasikan path final
			zap.ReplaceGlobals(logger)
			logger.Info("logger initialized", zap.String("path", logPath))
			return logger, nil
		}),

		// ✅ Provide *metrics.HTTPMetrics ke DI container
		metrics.Module(metrics.Config{Namespace: cfg.App.Name}),

		web.Module(cfg.Server.Port /* mounts via WireRoutes */),

		// ✅ Pasang CORS middleware dari config
		web.CORSModule(web.CORSConfig{
			Enabled:             cfg.Server.CORS.Enabled,
			Origins:             cfg.Server.CORS.Origins,
			Methods:             cfg.Server.CORS.Methods,
			Headers:             cfg.Server.CORS.Headers,
			ExposeHeaders:       cfg.Server.CORS.ExposeHeaders,
			Credentials:         cfg.Server.CORS.Credentials,
			MaxAgeSeconds:       cfg.Server.CORS.MaxAgeSeconds,
			AllowWildcard:       cfg.Server.CORS.AllowWildcard,
			AllowPrivateNetwork: cfg.Server.CORS.AllowPrivateNetwork,
		}),

		actuator.Module(cfg.Metrics.Prometheus.Path),

		data.Module(data.Config{
			DSN:              cfg.DB.DSN,
			MaxOpenConns:     cfg.DB.MaxOpenConns,
			MaxIdleConns:     cfg.DB.MaxIdleConns,
			ConnMaxLifetimeS: cfg.DB.ConnMaxLifetimeS,
		}),

		// ✅ Provide TxManager dari *gorm.DB
		fx.Provide(func(db *gorm.DB) data.TxManager {
			// Jika DB nil (auto-config off), tetap provide manager yang error saat dipakai
			return data.NewTxManager(db)
		}),

		// ✅ Provide Trace Middleware
		fx.Invoke(func(r *gin.Engine) {
			r.Use(web.TraceMiddleware())
		}),

		// Auto-migrate domain
		fx.Invoke(app.BootMigrate),

		// ✅ Pasang middleware metrics ke Gin
		fx.Invoke(func(r *gin.Engine, hm *metrics.HTTPMetrics) {
			r.Use(hm.Handler())
		}),

		// logger global
		fx.Invoke(func(l *zap.Logger) {
			zap.ReplaceGlobals(l)
			l.Info("logger initialized", zap.String("log_file", "/logs/app.log"))
		}),

		// Wire routes (public + protected)
		fx.Invoke(func(r *gin.Engine, db *gorm.DB, tx data.TxManager, logger *zap.Logger) {
			bootstrap.WireRoutes(
				r,
				db,
				security.JWTConfig{
					Enabled:  cfg.Security.Enabled,
					Secret:   cfg.Security.JWT.Secret,
					Issuer:   cfg.Security.JWT.Issuer,
					Audience: cfg.Security.JWT.Audience,
				},
				cfg.Security.JWT.TTLMin,
				tx,
				logger,
			)
		}),
	).Run()
}
