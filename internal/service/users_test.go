package service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
	sqlite3 "modernc.org/sqlite/lib"
)

type SQLiteUniqueErrMock struct{}

func (e *SQLiteUniqueErrMock) Code() int     { return sqlite3.SQLITE_CONSTRAINT_UNIQUE }
func (e *SQLiteUniqueErrMock) Error() string { return "" }

type CreateUserMock struct {
	QuerierMock
	DuplicateUsername string
	DuplicateEmail    string
}

func (m *CreateUserMock) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	m.Calls++
	if arg.Username == m.DuplicateUsername || arg.Email == m.DuplicateEmail {
		return db.User{}, &SQLiteUniqueErrMock{}
	}

	return db.User{}, nil
}

func TestAddUser(t *testing.T) {
	const (
		DuplicateUsername = "test"
		DuplicateEmail    = "test@example.com"
		Password          = "password"
		IsAdmin           = true
		BcryptLimit       = 73 // Max bytes bcrypt can handle
	)
	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		expectedError error
		expectedCalls int
	}{
		{"valid user request", fmt.Sprintf("%s1", DuplicateUsername), fmt.Sprintf("%s1", DuplicateEmail), Password, nil, 1},
		{"duplicate username", DuplicateUsername, fmt.Sprintf("%s1", DuplicateEmail), Password, service.ErrUserAlreadyExists, 1},
		{"duplicate email", fmt.Sprintf("%s1", DuplicateUsername), DuplicateEmail, Password, service.ErrUserAlreadyExists, 1},
		{"password too long", fmt.Sprintf("%s1", DuplicateUsername), fmt.Sprintf("%s1", DuplicateEmail), strings.Repeat("a", BcryptLimit), service.ErrPasswordHashCreation, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &CreateUserMock{DuplicateUsername: DuplicateUsername, DuplicateEmail: DuplicateEmail}
			service := service.NewService(mock)

			_, err := service.AddUser(t.Context(), tt.username, tt.email, tt.password, IsAdmin)
			AssertErrorsEqual(t, tt.expectedError, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}
