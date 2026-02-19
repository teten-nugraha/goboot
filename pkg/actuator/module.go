package actuator

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

func Module(metricsPath string) fx.Option {
	return fx.Invoke(func(r *gin.Engine) {
		if metricsPath == "" {
			metricsPath = "/actuator/metrics"
		}
		r.GET(metricsPath, gin.WrapH(promhttp.Handler()))
	})
}
