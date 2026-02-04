package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 3, 5, 7, 10},
	}, []string{"method", "path"})

	activeWebSocketConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_websocket_connections",
		Help: "Number of active WebSocket connections",
	})
)

func Metrics() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()

		status := c.Response().StatusCode()
		path := c.Path()
		method := c.Method()

		httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		return err
	}
}

func UpdateActiveWS(count int) {
	activeWebSocketConnections.Set(float64(count))
}
