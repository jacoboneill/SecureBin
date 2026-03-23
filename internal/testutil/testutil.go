package testutil

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jacoboneill/SecureBin/internal/db"
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
