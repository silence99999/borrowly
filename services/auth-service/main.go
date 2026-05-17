package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"

	_ "github.com/lib/pq"
)

type smtpConf struct {
	host, username, password, sender string
	port                             int
}

type config struct {
	addr string
	dsn  string
	smtp smtpConf
}

type application struct {
	cfg config
	db  *sql.DB
	wg  sync.WaitGroup
}

func main() {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	cfg := config{
		addr: ":5001",
		dsn:  os.Getenv("AUTH_DB_DSN"),
		smtp: smtpConf{
			host:     os.Getenv("SMTP_HOST"),
			port:     port,
			username: os.Getenv("SMTP_USERNAME"),
			password: os.Getenv("SMTP_PASSWORD"),
			sender:   os.Getenv("SMTP_SENDER"),
		},
	}

	db, err := sql.Open("postgres", cfg.dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("auth-service: db ping failed:", err)
	}
	defer db.Close()
	log.Println("auth-service: db connected")

	app := &application{cfg: cfg, db: db}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}