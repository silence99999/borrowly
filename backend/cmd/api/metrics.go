package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rent_items_http_requests_total",
			Help: "Total number of HTTP requests handled by the API.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rent_items_http_request_duration_seconds",
			Help:    "Duration of HTTP requests handled by the API.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)

	itemBrowseRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rent_items_item_browse_requests_total",
			Help: "Total number of item browse operations.",
		},
		[]string{"operation", "status"},
	)

	rentalCreationRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rent_items_rental_creation_requests_total",
			Help: "Total number of rental creation attempts.",
		},
		[]string{"status"},
	)
)

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *metricsResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (app *application) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func (app *application) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldSkipMetrics(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		writer := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(writer, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}

		status := strconv.Itoa(writer.statusCode)
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route, status).Observe(duration)
	})
}

func shouldSkipMetrics(path string) bool {
	switch path {
	case "/metrics", "/auth/me":
		return true
	default:
		return false
	}
}

func recordItemBrowseMetric(operation string, statusCode int) {
	status := "success"
	if statusCode >= 400 {
		status = "error"
	}

	itemBrowseRequestsTotal.WithLabelValues(operation, status).Inc()
}

func recordRentalCreationMetric(statusCode int) {
	status := "success"
	if statusCode >= 400 {
		status = "error"
	}

	rentalCreationRequestsTotal.WithLabelValues(status).Inc()
}
