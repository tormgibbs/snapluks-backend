package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/tormgibbs/snapluks-backend/internal/mailer"
	"github.com/tormgibbs/snapluks-backend/internal/s3"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/tormgibbs/snapluks-backend/internal/data"
	"github.com/tormgibbs/snapluks-backend/internal/jsonlog"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host   string
		port   int
		user   string
		pass   string
		sender string
	}
}

type application struct {
	config   config
	mailer   mailer.Mailer
	models   data.Models
	wg       sync.WaitGroup
	logger   *jsonlog.Logger
	s3Client *s3.Client
}

func main() {
	cfg := loadConfig()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	s3Client, err := s3.NewClient("snapluks")
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config:   cfg,
		logger:   logger,
		models:   data.NewModels(db),
		mailer:   mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.user, cfg.smtp.pass, cfg.smtp.sender),
		s3Client: s3Client,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func loadConfig() config {
	_ = godotenv.Load()

	configErrors := make([]error, 0)

	getEnv := func(key, fallback string) string {
		val := os.Getenv(key)
		if val == "" {
			return fallback
		}
		return val
	}

	mustGetEnv := func(key string) string {
		val := os.Getenv(key)
		if val == "" {
			configErrors = append(configErrors, fmt.Errorf("environment variable %s is required but not set", key))
			return ""
		}
		return val
	}

	atoi := func(s string, fallback int) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			return fallback
		}
		return i
	}

	cfg := config{}
	cfg.port = atoi(getEnv("PORT", "4000"), 4000)
	cfg.env = getEnv("ENV", "development")

	cfg.db.dsn = mustGetEnv("DB_DSN")
	cfg.db.maxOpenConns = atoi(getEnv("DB_MAX_OPEN_CONNS", "25"), 25)
	cfg.db.maxIdleConns = atoi(getEnv("DB_MAX_IDLE_CONNS", "25"), 25)
	cfg.db.maxIdleTime = getEnv("DB_MAX_IDLE_TIME", "15m")

	cfg.smtp.host = mustGetEnv("SMTP_HOST")
	cfg.smtp.port = atoi(getEnv("SMTP_PORT", "587"), 587)
	cfg.smtp.user = mustGetEnv("SMTP_USER")
	cfg.smtp.pass = mustGetEnv("SMTP_PASS")
	cfg.smtp.sender = mustGetEnv("SMTP_SENDER")

	if len(configErrors) > 0 {
		fmt.Println(errors.Join(configErrors...))
		os.Exit(1)
	}

	return cfg
}
