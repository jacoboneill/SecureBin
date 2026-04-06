package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
)

type GetSessionMock struct {
	QuerierMock
	ValidSession *db.Session
}

type CreateSessionMock struct {
	QuerierMock
	ValidUserID int64
}

type DeleteSessionMock struct {
	QuerierMock
	ValidSessionID string
}

func (m *GetSessionMock) GetSession(ctx context.Context, id string) (db.Session, error) {
	m.Calls++
	if id != m.ValidSession.ID {
		return db.Session{}, sql.ErrNoRows
	}

	return *m.ValidSession, nil
}

func (m *CreateSessionMock) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
	m.Calls++
	if arg.UserID != m.ValidUserID {
		return db.Session{}, ErrDBMock
	}

	return db.Session{}, nil
}

func (m *DeleteSessionMock) DeleteSession(ctx context.Context, id string) (sql.Result, error) {
	m.Calls++
	if id != m.ValidSessionID {
		return ResultMock{0}, nil
	}

	return ResultMock{1}, nil
}

func TestValidateSession(t *testing.T) {
	validSession := &db.Session{
		ID:        "123",
		UserID:    0,
		CreatedAt: time.Time{},
	}
	tests := []struct {
		name          string
		sessionID     string
		expectedErr   error
		expectedCalls int
	}{
		{"valid session ID", validSession.ID, nil, 1},
		{"invalid session ID", Modify(validSession.ID), service.ErrSessionNotFound, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GetSessionMock{ValidSession: validSession}
			service := service.NewService(mock)

			err := service.ValidateSession(t.Context(), tt.sessionID)
			AssertErrorsEqual(t, tt.expectedErr, err)
			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}

func TestCreateSession(t *testing.T) {
	const validUserID = 1
	tests := []struct {
		name          string
		userID        int64
		expectedErr   error
		expectedCalls int
	}{
		{"valid user ID", validUserID, nil, 1},
		{"invalid user ID", validUserID + 1, service.ErrSessionCreationFailed, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &CreateSessionMock{ValidUserID: validUserID}
			service := service.NewService(mock)

			_, err := service.CreateSession(t.Context(), tt.userID)
			AssertErrorsEqual(t, tt.expectedErr, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	const validSessionID = "123"
	tests := []struct {
		name          string
		sessionID     string
		expectedErr   error
		expectedCalls int
	}{
		{"valid session ID", validSessionID, nil, 1},
		{"invalid session ID", Modify(validSessionID), service.ErrSessionNotFound, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeleteSessionMock{ValidSessionID: validSessionID}
			service := service.NewService(mock)

			err := service.DeleteSession(t.Context(), tt.sessionID)
			AssertErrorsEqual(t, tt.expectedErr, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}
