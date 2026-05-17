package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type contextKey string

const authUserKey contextKey = "authUser"

type AuthUser struct {
	ID   string
	Role string
}

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rental_service_http_requests_total",
		Help: "Total HTTP requests handled by rental-service.",
	}, []string{"method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rental_service_http_request_duration_seconds",
		Help:    "Duration of HTTP requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status"})

	rentalCreations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rental_service_rental_creation_total",
		Help: "Total rental creation attempts.",
	}, []string{"status"})
)

type metricsRW struct {
	http.ResponseWriter
	code int
}

func (m *metricsRW) WriteHeader(code int) {
	m.code = code
	m.ResponseWriter.WriteHeader(code)
}

func (app *application) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rw := &metricsRW{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}
		status := strconv.Itoa(rw.code)
		dur := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, route, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, route, status).Observe(dur)
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth_token")
		if err != nil {
			writeError(w, http.StatusUnauthorized, errors.New("missing auth cookie"))
			return
		}
		tok, err := jwt.Parse(c.Value, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(os.Getenv("SECRET")), nil
		})
		if err != nil || !tok.Valid {
			writeError(w, http.StatusUnauthorized, errors.New("invalid token"))
			return
		}
		claims, _ := tok.Claims.(jwt.MapClaims)
		exp, _ := claims["exp"].(float64)
		if time.Now().Unix() > int64(exp) {
			writeError(w, http.StatusUnauthorized, errors.New("token expired"))
			return
		}
		sub, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		ctx := context.WithValue(r.Context(), authUserKey, AuthUser{ID: sub, Role: role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
