package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type smtpConf struct {
	host, username, password, sender string
	port                             int
}

type config struct {
	addr           string
	dsn            string
	smtp           smtpConf
	itemServiceURL string
	authServiceURL string
}

type application struct {
	cfg    config
	db     *sql.DB
	client *http.Client
}

func main() {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	cfg := config{
		addr:           ":5003",
		dsn:            os.Getenv("RENTAL_DB_DSN"),
		itemServiceURL: os.Getenv("ITEM_SERVICE_URL"),
		authServiceURL: os.Getenv("AUTH_SERVICE_URL"),
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
		log.Fatal("rental-service: db ping failed:", err)
	}
	defer db.Close()
	log.Println("rental-service: db connected")

	app := &application{
		cfg:    cfg,
		db:     db,
		client: &http.Client{Timeout: 5 * time.Second},
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	app.startReminderWorker()

	if err := app.serve(); err != nil {
		log.Fatal(err)
	}
}
