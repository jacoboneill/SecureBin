package testutil

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jacoboneill/SecureBin/internal/db"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func SetupTestDB(t *testing.T) (*db.Queries, *sql.DB) {
	t.Helper()

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	driver, err := sqlite.WithInstance(conn, &sqlite.Config{})
	if err != nil {
		t.Fatal(err)
	}

	source, err := iofs.New(db.Migrations, "migrations")
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	return db.New(conn), conn
}

func SeedUser(t *testing.T, q *db.Queries) db.User {
	const defaultPassword = "password"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("bcrypt failed to hash password")
	}

	user, err := q.RegisterUser(t.Context(), db.RegisterUserParams{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		IsAdmin:      true,
	})
	if err != nil {
		t.Fatalf("error on user creation %q", err)
	}

	return user
}
