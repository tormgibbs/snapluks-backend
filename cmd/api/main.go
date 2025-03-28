package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
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
}

type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	dsn := getDSN()

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(
		&cfg.env, "env", "development", "Environment (development|staging|production)",
	)
	flag.StringVar(
		&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN",
	)
	flag.IntVar(
		&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections",
	)
	flag.IntVar(
		&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections",
	)
	flag.StringVar(
		&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time",
	)
	flag.Parse()

	logger := log.New(os.Stdout, "api: ", log.LstdFlags|log.Lshortfile)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	logger.Println("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting server on port %d", cfg.port)
	err = srv.ListenAndServe()
	logger.Fatal(err)
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

func getDSN() string {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		panic(fmt.Errorf("DB_DSN missing from .env"))
	}
	return dsn
}
