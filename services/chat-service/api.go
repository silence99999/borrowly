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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Group(func(p chi.Router) {
		p.Use(requireAuth)
		p.Get("/chats", app.listConversations)
		p.Get("/chats/{userID}", app.getMessages)
		p.Post("/chats/{userID}", app.sendMessage)
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
		log.Println("chat-service: shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ch <- srv.Shutdown(ctx)
	}()

	log.Printf("chat-service: listening on %s", app.cfg.addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return <-ch
}
