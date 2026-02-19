package schedule

import (
	"context"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func Module(jobs ...func()) fx.Option {
	return fx.Invoke(func(lc fx.Lifecycle) {
		c := cron.New()
		for _, job := range jobs {
			// contoh: tiap 1 menit
			_, _ = c.AddFunc("@every 1m", job)
		}
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error { c.Start(); return nil },
			OnStop:  func(ctx context.Context) error { c.Stop(); return nil },
		})
	})
}
