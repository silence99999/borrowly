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
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(app.MetricsMiddleware)

	r.Handle("/metrics", app.MetricsHandler())

	r.Post("/auth/signup", app.SignUp)
	r.Post("/auth/login", app.Login)
	r.Post("/auth/verify-email", app.VerifyEmail)
	r.Post("/auth/logout", app.Logout)

	r.Get("/items", app.GetItemsHandler)
	r.Get("/items/{id}", app.GetItemByIDHandler)
	r.Get("/pickup-points", app.GetPickupPoints)

	r.Group(func(protected chi.Router) {
		protected.Use(app.RequireAuth)

		protected.Get("/auth/me", app.Me)
		protected.Get("/items/my", app.GetMyItemsHandler)
		protected.Post("/rentals", app.CreateRentalHandler)
		protected.Get("/rentals/my", app.GetMyRentalsHandler)

		protected.Patch("/rental/{id}/cancel", app.CancelRentalHandler)
		protected.Patch("/rental/{id}/pay", app.PayRentalHandler)

		protected.Post("/items", app.CreateItemHandler)

		protected.Patch("/items/{id}", app.UpdateItemHandler)

		protected.Delete("/items/{id}", app.DeleteItemHandler)

		protected.Post("/items/{id}/reviews", app.CreateReviewHandler)
		protected.Get("/items/{id}/reviews", app.GetReviewsHandler)
		protected.Delete("/items/{id}/reviews/{reviewID}", app.DeleteReviewHandler)
		protected.Patch("/items/{id}/reviews/{reviewID}", app.UpdateReviewHandler)

		protected.Post("/items/{id}/images", app.CreateItemImageHandler)

		protected.With(app.RequireAdmin).Post("/pickup-points", app.CreatePickupPoint)
	})

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:    app.config.addr,
		Handler: h,
	}

	shutdownError := make(chan error)

	go func() {

		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		log.Println("shutting down server", map[string]string{
			"signal": s.String(),
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		log.Println("completing background tasks")

		shutdownError <- nil
	}()

	log.Printf("Server has started at address %s", app.config.addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownError
	if err != nil {
		return err
	}

	log.Println("Server has stopped")

	return nil
}
