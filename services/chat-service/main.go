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
	cfg            config
	db             *sql.DB
	authServiceURL string
}

func main() {
	cfg := config{
		addr: ":5005",
		dsn:  os.Getenv("CHAT_DB_DSN"),
	}

	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("chat-service: db ping failed:", err)
	}
	defer db.Close()
	log.Println("chat-service: db connected")

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://auth-service:5001"
	}
	app := &application{cfg: cfg, db: db, authServiceURL: authURL}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
