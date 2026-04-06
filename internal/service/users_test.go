package service_test

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
	"golang.org/x/crypto/bcrypt"
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

type GetUserByEmailOrUsernameMock struct {
	QuerierMock
	ValidUser *db.User
}

type GetUserMock struct {
	QuerierMock
	ValidUser *db.User
}

type GetUserFromSessionMock struct {
	QuerierMock
	ValidUser    *db.User
	ValidSession *db.Session
}

func (m *CreateUserMock) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	m.Calls++
	if arg.Username == m.DuplicateUsername || arg.Email == m.DuplicateEmail {
		return db.User{}, &SQLiteUniqueErrMock{}
	}

	return db.User{}, nil
}

func (m *GetUserByEmailOrUsernameMock) GetUserByEmailOrUsername(ctx context.Context, identifier string) (db.User, error) {
	m.Calls++
	if identifier != m.ValidUser.Email && identifier != m.ValidUser.Username {
		return db.User{}, sql.ErrNoRows
	}

	return *m.ValidUser, nil
}

func (m *GetUserMock) GetUser(ctx context.Context, id int64) (db.User, error) {
	m.Calls++
	if id == m.ValidUser.ID {
		return *m.ValidUser, nil
	}
	return db.User{}, sql.ErrNoRows
}

func (m *GetUserFromSessionMock) GetSession(ctx context.Context, id string) (db.Session, error) {
	m.Calls++
	return (&GetSessionMock{ValidSession: m.ValidSession}).GetSession(ctx, id)
}

func (m *GetUserFromSessionMock) GetUser(ctx context.Context, id int64) (db.User, error) {
	m.Calls++
	return (&GetUserMock{ValidUser: m.ValidUser}).GetUser(ctx, id)
}

func TestAddUser(t *testing.T) {
	const (
		duplicateUsername = "test"
		duplicateEmail    = "test@example.com"
		password          = "password"
		isAdmin           = true
		bcryptLimit       = 73 // Max bytes bcrypt can handle
	)
	tests := []struct {
		name          string
		username      string
		email         string
		password      string
		expectedError error
		expectedCalls int
	}{
		{"valid user request", Modify(duplicateUsername), Modify(duplicateEmail), password, nil, 1},
		{"duplicate username", duplicateUsername, Modify(duplicateEmail), password, service.ErrUserAlreadyExists, 1},
		{"duplicate email", Modify(duplicateUsername), duplicateEmail, password, service.ErrUserAlreadyExists, 1},
		{"password too long", Modify(duplicateUsername), Modify(duplicateEmail), strings.Repeat("a", bcryptLimit), service.ErrPasswordHashCreation, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &CreateUserMock{DuplicateUsername: duplicateUsername, DuplicateEmail: duplicateEmail}
			service := service.NewService(mock)

			_, err := service.AddUser(t.Context(), tt.username, tt.email, tt.password, isAdmin)
			AssertErrorsEqual(t, tt.expectedError, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}

func AssertUser(t testing.TB, expectedUser, capturedUser *db.User) {
	t.Helper()
	if !reflect.DeepEqual(capturedUser, expectedUser) {
		t.Errorf("expected %+v, got %+v", expectedUser, capturedUser)
	}
}

func TestAuthenticateUser(t *testing.T) {
	const validPassword = "password"

	validPasswordHash, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	validUser := &db.User{
		ID:           0,
		Username:     "test",
		Email:        "test@example.com",
		PasswordHash: string(validPasswordHash),
		IsAdmin:      true,
		CreatedAt:    time.Time{},
	}

	tests := []struct {
		name          string
		username      string
		password      string
		expectedError error
		expectedCalls int
	}{
		{"valid username and password", validUser.Username, validPassword, nil, 1},
		{"valid email and password", validUser.Email, validPassword, nil, 1},
		{"invalid username, valid password", Modify(validUser.Username), validPassword, service.ErrUserNotFound, 1},
		{"valid username, invalid password", validUser.Username, Modify(validPassword), service.ErrInvalidPassword, 1},
		{"invalid username, invalid password", Modify(validUser.Username), Modify(validPassword), service.ErrUserNotFound, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GetUserByEmailOrUsernameMock{ValidUser: validUser}
			service := service.NewService(mock)

			user, err := service.AuthenticateUser(t.Context(), tt.username, tt.password)
			AssertErrorsEqual(t, tt.expectedError, err)
			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
			if err == nil {
				AssertUser(t, user, validUser)
			}
		})
	}
}

func TestGetUserFromSession(t *testing.T) {
	var (
		validUser = db.User{
			ID:           1,
			Username:     "test",
			Email:        "test@example.com",
			PasswordHash: "",
			IsAdmin:      false,
			CreatedAt:    time.Time{},
		}
		validSession = db.Session{
			ID:        "123",
			UserID:    validUser.ID,
			CreatedAt: time.Time{},
		}
	)

	orphanedSession := validSession
	orphanedSession.UserID = validUser.ID + 1

	tests := []struct {
		name          string
		sessionID     string
		mockSession   *db.Session
		expectedError error
		expectedCalls int
	}{
		{"valid session", validSession.ID, &validSession, nil, 2},
		{"invalid session", Modify(validSession.ID), &validSession, service.ErrSessionNotFound, 1},
		{"orphaned session", orphanedSession.ID, &orphanedSession, service.ErrUserNotFound, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GetUserFromSessionMock{ValidSession: tt.mockSession, ValidUser: &validUser}
			service := service.NewService(mock)

			user, err := service.GetUserFromSession(t.Context(), tt.sessionID)
			AssertErrorsEqual(t, tt.expectedError, err)
			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
			if err == nil {
				AssertUser(t, user, &validUser)
			}
		})
	}
}

func TestGetUserFromContext(t *testing.T) {
	tests := []struct {
		name             string
		user             *db.User
		isUserIDInCtxKey bool
		expectedError    error
	}{
		{"user from context", &db.User{ID: 0, Username: "test", Email: "test@example.com", PasswordHash: "", IsAdmin: false, CreatedAt: time.Time{}}, true, nil},
		{"invalid user from context", &db.User{}, false, service.ErrUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			if tt.isUserIDInCtxKey {
				ctx = context.WithValue(ctx, contextkey.UserCtxKey, tt.user)
			}

			service := service.NewService(struct{ db.Querier }{})
			user, err := service.GetUserFromContext(ctx)
			AssertErrorsEqual(t, tt.expectedError, err)
			if err == nil {
				AssertUser(t, tt.user, user)
			}
		})
	}
}
