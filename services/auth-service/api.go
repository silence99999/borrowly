package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Post("/auth/signup", app.signUp)
	r.Post("/auth/login", app.login)
	r.Post("/auth/verify-email", app.verifyEmail)
	r.Post("/auth/logout", app.logout)

	// Internal service-to-service endpoint — not exposed via gateway
	r.Get("/internal/users/{id}", app.internalGetUser)

	r.Group(func(p chi.Router) {
		p.Use(requireAuth)
		p.Get("/auth/me", app.me)
	})

	return r
}

func (app *application) serve() error {
	srv := &http.Server{Addr: app.cfg.addr, Handler: app.routes()}
	ch := make(chan error, 1)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("auth-service: shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ch <- srv.Shutdown(ctx)
	}()

	log.Printf("auth-service: listening on %s", app.cfg.addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return <-ch
}
