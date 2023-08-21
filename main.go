package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var (
	buildVersion string
	version      string
)

type config struct {
	Addr            string
	BaseURI         string
	ShutdownTimeout time.Duration
	DBConnString    string
	DBMigrations    string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg := config{}
	flag.StringVar(&cfg.Addr, "addr", ":9001", "Address to listen on")
	flag.StringVar(&cfg.BaseURI, "base-uri", "", "Base URI")
	flag.DurationVar(&cfg.ShutdownTimeout, "shutdown-timeout", 2*time.Second, "Shutdown timeout")
	flag.StringVar(&cfg.DBConnString, "db-conn", "", "DB connection string")
	flag.StringVar(&cfg.DBMigrations, "db-migrations", "db/migrations", "DB migrations path")
	var flagVersion, flagBuildVersion bool
	flag.BoolVar(&flagVersion, "version", false, "Print version")
	flag.BoolVar(&flagBuildVersion, "build-version", false, "Print build version")
	flag.Parse()

	if flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flagBuildVersion {
		fmt.Println(buildVersion)
		os.Exit(0)
	}

	if cfg.DBConnString == "" {
		if s := os.Getenv("DB_CONN"); s != "" {
			cfg.DBConnString = s
		} else {
			logger.Error("DB connection string is required")
			os.Exit(1)
		}
	}

	if err := run(logger, cfg); err != nil {
		logger.Error("Error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, cfg config) error {
	// Run DB migrations.
	logger.Info("Running DB migrations")
	migrations, err := migrate.New("file://"+cfg.DBMigrations, cfg.DBConnString)
	if err != nil {
		return fmt.Errorf("Failed to initialize DB migrations: %w", err)
	}

	err = migrations.Up()
	switch err {
	case nil:
	case migrate.ErrNoChange:
		logger.Info("No DB schema changes")
	default:
		return fmt.Errorf("Failed to run DB migrations: %w", err)
	}

	// Initialize a DB connection pool.
	logger.Info("Initializing DB connection pool")
	db, err := sql.Open("postgres", cfg.DBConnString)
	if err != nil {
		return fmt.Errorf("Failed to open DB connection: %w", err)
	}
	defer db.Close()

	// Check DB connection.
	// TODO: Add retry logic with exponential backoff.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("Failed to ping DB: %w", err)
	}

	middlewares := []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	}

	rtr := chi.NewRouter()
	rtr.Use(middlewares...)
	if cfg.BaseURI == "" {
		cfg.BaseURI = "/"
	}
	rtr.Get(cfg.BaseURI, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Not implemented"))
	}))

	// Configure HTTP server.
	// Timeouts can/should be individually fine tuned.
	// Here we'll use the same value.
	httpTimeout := 30 * time.Second
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           rtr,
		IdleTimeout:       httpTimeout,
		ReadHeaderTimeout: httpTimeout,
		ReadTimeout:       httpTimeout,
		WriteTimeout:      httpTimeout,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		// Received an interrupt signal, shutting down.
		logger.Info("Shutting down service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		// Drain and close http connections.
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Shutdown error: %s", err)
		}
		close(idleConnsClosed)
	}()

	logger.Info("Starting service", "version", version, "baseURI", cfg.BaseURI, "addr", cfg.Addr)
	// ListenAndServe always return a non-nil error.
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Don't wait, just return with the error.
		return err
	}

	// Wait until we shut down the server.
	<-idleConnsClosed
	logger.Info("Service stopped")

	return nil
}
