package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/jacoboneill/SecureBin/internal/db"
)

var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionCreationFailed = errors.New("failed to add new session to database")
)

func (s *Service) ValidateSession(ctx context.Context, sessionID string) error {
	_, err := s.queries.GetSession(ctx, sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSessionNotFound
	}
	return err
}

func (s *Service) CreateSession(ctx context.Context, userID int64) (string, error) {
	token := make([]byte, 32)
	rand.Read(token)
	sessionID := base64.URLEncoding.EncodeToString(token)

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{ID: sessionID, UserID: userID})
	if err != nil {
		return "", ErrSessionCreationFailed
	}

	return session.ID, nil
}

func (s *Service) DeleteSession(ctx context.Context, sessionID string) error {
	return s.queries.DeleteSession(ctx, sessionID)
}
