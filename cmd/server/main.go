package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	db "github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/handlers"
	_ "modernc.org/sqlite"
)

func runMigrations(conn *sql.DB, sourceURL string) error {
	driver, err := sqlite.WithInstance(conn, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "sqlite", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func main() {
	const port = 8080
	const migrationSourceURL = "file://internal/db/migrations"
	const dbName = "securebin.db"

	conn, err := sql.Open("sqlite", dbName)
	if err != nil {
		slog.Error("server failed to connect to database", "err", err)
	}
	slog.Info("database connection secured", "db", dbName)
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("server failed to close database", "err", err)
		}
	}()

	if err := runMigrations(conn, migrationSourceURL); err != nil {
		slog.Error("migrations failed to apply", "err", err)
	}
	slog.Info("migrations applied", "sourceURL", migrationSourceURL)

	h := handlers.New(db.New(conn))
	mux := h.NewRouter()
	slog.Info("router initialised")

	slog.Info("server started", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
