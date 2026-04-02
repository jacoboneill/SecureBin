package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/db"
	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrPasswordHashCreation = errors.New("password hash failed")
	ErrUserCreation         = errors.New("failed to create user")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidPassword      = errors.New("username and password do not match")
)

func (s *Service) AddUser(ctx context.Context, username, email, password string, isAdmin bool) (*db.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrPasswordHashCreation, err)
	}

	user, err := s.queries.CreateUser(ctx, db.CreateUserParams{Username: username, Email: email, PasswordHash: string(passwordHash), IsAdmin: isAdmin})
	if err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("%w: %w", ErrUserCreation, err)
	}
	return &user, nil
}

func (s *Service) AuthenticateUser(ctx context.Context, username, password string) (*db.User, error) {
	user, err := s.queries.GetUserByEmailOrUsername(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		} else {
			return nil, err
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return &user, nil
}

func (s *Service) GetUserFromSession(ctx context.Context, sessionID string) (*db.User, error) {
	session, err := s.queries.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	user, err := s.queries.GetUser(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *Service) GetUserFromContext(ctx context.Context) (*db.User, error) {
	user, ok := ctx.Value(contextkey.UserCtxKey).(*db.User)
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}
