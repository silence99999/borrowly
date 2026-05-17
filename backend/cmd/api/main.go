package main

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/silence99999/advanced_final/internal/data"
	"github.com/silence99999/advanced_final/internal/mailer"
)

type application struct {
	config config
	models data.Models
	wg     sync.WaitGroup
	mailer mailer.Mailer
}

type config struct {
	addr string
	db   dbconfig
	smtp smtpConfig
}

type smtpConfig struct {
	host     string
	port     int
	username string
	password string
	sender   string
}

type dbconfig struct {
	dsn string
}

func main() {

	dsn := os.Getenv("DB_DSN")

	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))

	if err != nil {
		slog.Error("wrong smtp port", err)
		os.Exit(1)
	}

	cfg := config{
		addr: ":5000",
		db: dbconfig{
			dsn: dsn,
		},
		smtp: smtpConfig{
			os.Getenv("SMTP_HOST"),
			smtpPort,
			os.Getenv("SMTP_USERNAME"),
			os.Getenv("SMTP_PASSWORD"),
			os.Getenv("SMTP_SENDER"),
		},
	}

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Successfully connected to db")

	defer db.Close()

	api := &application{
		models: data.NewModels(db),
		config: cfg,
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	api.startRentalReminderWorker()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	if err := api.run(api.routes()); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
