package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
)

type config struct {
	addr string
	dsn  string
}

type application struct {
	cfg config
	db  *sql.DB
}

func main() {
	cfg := config{
		addr: ":5004",
		dsn:  os.Getenv("REVIEW_DB_DSN"),
	}

	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("review-service: db ping failed:", err)
	}
	defer db.Close()
	log.Println("review-service: db connected")

	app := &application{cfg: cfg, db: db}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
