package prometheus

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var requestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total requests",
	},
	[]string{"service", "endpoint", "method", "status"},
)

var requestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Request latency",
		Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1, 2},
	},
	[]string{"service", "endpoint"},
)

var OrdersCompleted = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "order_completed_total",
		Help: "Total completed orders",
	},
)

var OrdersFailed = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "order_failed_total",
		Help: "Total failed orders",
	},
)

var ProductInventory = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "product_inventory",
		Help: "Current inventory levels",
	},
	[]string{"product_id"},
)

func Register() {
	prometheus.MustRegister(
		requestCount,
		requestDuration,
		OrdersCompleted,
		OrdersFailed,
		ProductInventory,
	)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Instrument(handler http.HandlerFunc, service, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: 200}

		handler(rec, r)

		duration := time.Since(start).Seconds()

		requestCount.WithLabelValues(
			service,
			endpoint,
			r.Method,
			strconv.Itoa(rec.status),
		).Inc()

		requestDuration.WithLabelValues(service, endpoint).
			Observe(duration)
	}
}
