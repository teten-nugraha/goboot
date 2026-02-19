// pkg/metrics/http_metrics.go
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Label konvensi: selaras dengan best-practice Prometheus
	labelMethod = "method"
	labelPath   = "path"
	labelStatus = "status"
	labelHost   = "host"
)

// HTTPMetrics menyimpan metrik2 HTTP dan expose middleware Gin.
type HTTPMetrics struct {
	namespace string

	reqTotal    *prometheus.CounterVec
	reqDuration *prometheus.HistogramVec
	reqSize     *prometheus.SummaryVec
	respSize    *prometheus.SummaryVec
}

// NewHTTPMetrics mendaftarkan metriks ke default registry.
func NewHTTPMetrics(namespace string) (*HTTPMetrics, error) {
	if namespace == "" {
		namespace = "app"
	}

	m := &HTTPMetrics{
		namespace: namespace,
		reqTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Jumlah total HTTP request yang diproses.",
			},
			[]string{labelMethod, labelPath, labelStatus, labelHost},
		),
		reqDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Durasi HTTP request (detik).",
				// Bucket mirip default client_golang
				Buckets: prometheus.DefBuckets,
			},
			[]string{labelMethod, labelPath, labelStatus, labelHost},
		),
		reqSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  "http",
				Name:       "request_size_bytes",
				Help:       "Ukuran payload request (bytes).",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{labelMethod, labelPath, labelHost},
		),
		respSize: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  namespace,
				Subsystem:  "http",
				Name:       "response_size_bytes",
				Help:       "Ukuran payload response (bytes).",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{labelMethod, labelPath, labelStatus, labelHost},
		),
	}

	// Register ke default registry
	prometheus.MustRegister(m.reqTotal, m.reqDuration, m.reqSize, m.respSize)

	return m, nil
}

// Handler mengembalikan middleware Gin untuk observasi metrik.
func (m *HTTPMetrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqSize := approximateRequestSize(c.Request)

		c.Next() // proses request

		route := c.FullPath()
		if route == "" {
			// fallback ke URL path mentah jika route belum dikenali oleh Gin (mis. 404)
			route = c.Request.URL.Path
		}
		labels3 := prometheus.Labels{
			labelMethod: c.Request.Method,
			labelPath:   route,
			labelHost:   c.Request.Host,
		}
		statusCode := c.Writer.Status()
		labels4 := prometheus.Labels{
			labelMethod: c.Request.Method,
			labelPath:   route,
			labelStatus: strconv.Itoa(statusCode),
			labelHost:   c.Request.Host,
		}

		duration := time.Since(start).Seconds()
		respSize := float64(c.Writer.Size())

		m.reqTotal.With(labels4).Inc()
		m.reqDuration.With(labels4).Observe(duration)
		m.reqSize.With(labels3).Observe(float64(reqSize))
		if respSize >= 0 {
			m.respSize.With(labels4).Observe(respSize)
		}
	}
}

func approximateRequestSize(r *http.Request) int {
	// Estimasi cukup untuk metrik (tidak harus presisi).
	s := 0
	if r.URL != nil {
		s += len(r.URL.Path)
	}
	s += len(r.Method)
	s += len(r.Proto)
	for k, v := range r.Header {
		s += len(k)
		for _, vv := range v {
			s += len(vv)
		}
	}
	if r.Host != "" {
		s += len(r.Host)
	}
	// Body length kalau ada Content-Length
	if r.ContentLength > 0 {
		s += int(r.ContentLength)
	}
	return s
}
