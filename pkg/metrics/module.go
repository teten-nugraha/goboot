// pkg/metrics/module.go
package metrics

import (
	"log"

	"go.uber.org/fx"
)

// Config minimal—pakai nama aplikasi sebagai namespace metrik.
type Config struct {
	Namespace string
}

// Module mendaftarkan HTTP metrics dan expose sebagai dependency Fx.
func Module(cfg Config) fx.Option {
	if cfg.Namespace == "" {
		cfg.Namespace = "app"
	}
	return fx.Options(
		fx.Provide(func() (*HTTPMetrics, error) {
			m, err := NewHTTPMetrics(cfg.Namespace)
			if err != nil {
				return nil, err
			}
			log.Printf("[metrics] prometheus http metrics initialized (namespace=%s)\n", cfg.Namespace)
			return m, nil
		}),
	)
}
