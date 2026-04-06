package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	db "github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/handler"
	"github.com/jacoboneill/SecureBin/internal/service"
	"golang.org/x/crypto/bcrypt"
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

// HACK: Remove when user registration is implemented
func seed(queries *db.Queries) {
	const defaultPassword = "password"

	// Create admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("bcrypt failed to hash password", "err", err)
	}
	queries.CreateUser(context.Background(), db.CreateUserParams{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		IsAdmin:      true,
	})
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

	queries := db.New(conn)
	h := handler.NewHandler(service.NewService(queries))
	mux := h.NewRouter()
	slog.Info("router initialised")

	// HACK: Remove when user registration is implemented
	seed(queries)
	slog.Info("new user added", "users", []struct {
		Username string
		Email    string
		Password string
	}{
		{"admin", "admin@example.com", "password"},
	})

	slog.Info("server started", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
