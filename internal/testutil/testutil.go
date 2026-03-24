package testutil

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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

type RegisterUserParams struct {
	Username string
	Email    string
	Password string
	IsAdmin  bool
}

type User struct {
	db.User
	SessionID string
}

func SeedUser(t *testing.T, q *db.Queries, registerUserParams RegisterUserParams) User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerUserParams.Password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal("bcrypt failed to hash password")
	}

	user, err := q.RegisterUser(t.Context(), db.RegisterUserParams{
		Username:     registerUserParams.Username,
		Email:        registerUserParams.Email,
		PasswordHash: string(hashedPassword),
		IsAdmin:      registerUserParams.IsAdmin,
	})
	if err != nil {
		t.Fatalf("error on user creation %q", err)
	}

	token := make([]byte, 32)
	rand.Read(token)
	sessionID := base64.URLEncoding.EncodeToString(token)

	session, err := q.CreateSession(t.Context(), db.CreateSessionParams{ID: sessionID, UserID: user.ID})
	if err != nil {
		t.Fatalf("error on session creation %q", err)
	}

	return User{
		User:      user,
		SessionID: session.ID,
	}
}
