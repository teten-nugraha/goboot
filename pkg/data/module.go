package data

import (
	"time"

	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN              string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetimeS int
}

func Module(cfg Config) fx.Option {
	return fx.Options(
		fx.Provide(func() (*gorm.DB, error) {
			if cfg.DSN == "" {
				return nil, nil
			}
			db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
			if err != nil {
				return nil, err
			}
			sqlDB, _ := db.DB()
			if cfg.MaxOpenConns > 0 {
				sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
			}
			if cfg.MaxIdleConns > 0 {
				sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
			}
			if cfg.ConnMaxLifetimeS > 0 {
				sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeS) * time.Second)
			}
			return db, nil
		}),
	)
}
