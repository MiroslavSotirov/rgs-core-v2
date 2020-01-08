package metrics

//import (
//	"github.com/prometheus/client_golang/prometheus"
//	"github.com/prometheus/client_golang/prometheus/promhttp"
//)
//
//var (
//	counter = prometheus.NewCounter(
//		prometheus.CounterOpts{
//			Namespace: "golang",
//			Name:      "my_counter",
//			Help:      "This is my counter",
//		})
//
//	gauge = prometheus.NewGauge(
//		prometheus.GaugeOpts{
//			Namespace: "golang",
//			Name:      "my_gauge",
//			Help:      "This is my gauge",
//		})
//
//	histogram = prometheus.NewHistogram(
//		prometheus.HistogramOpts{
//			Namespace: "golang",
//			Name:      "my_histogram",
//			Help:      "This is my histogram",
//		})
//
//	summary = prometheus.NewSummary(
//		prometheus.SummaryOpts{
//			Namespace: "golang",
//			Name:      "my_summary",
//			Help:      "This is my summary",
//		})
//)

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dflBuckets = []float64{300, 1200, 5000}
)

const (
	reqsName    = "rgs_api_requests_total"
	latencyName = "rgs_api_request_duration_milliseconds"
)

// Middleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type Middleware struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewMiddleware returns a new prometheus Middleware handler.
func NewHTTPMiddleware(name string, buckets ...float64) func(next http.Handler) http.Handler {
	var m Middleware
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)

	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)
	return m.handler
}

func (c Middleware) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		c.reqs.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Inc()
		c.latency.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
	}
	return http.HandlerFunc(fn)
}
