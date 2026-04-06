package service_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
)

type GetSessionMock struct {
	QuerierMock
	AvailableSessionID string
}

type CreateSessionMock struct {
	QuerierMock
	AvailableUserID int64
}

type DeleteSessionMock struct {
	QuerierMock
	AvailableSessionID string
}

func (m *GetSessionMock) GetSession(ctx context.Context, id string) (db.Session, error) {
	m.Calls++
	if id != m.AvailableSessionID {
		return db.Session{}, sql.ErrNoRows
	}

	return db.Session{}, nil
}

func (m *CreateSessionMock) CreateSession(ctx context.Context, arg db.CreateSessionParams) (db.Session, error) {
	m.Calls++
	if arg.UserID != m.AvailableUserID {
		return db.Session{}, ErrDBMock
	}

	return db.Session{}, nil
}

func (m *DeleteSessionMock) DeleteSession(ctx context.Context, id string) (sql.Result, error) {
	m.Calls++
	if id != m.AvailableSessionID {
		return ResultMock{0}, nil
	}

	return ResultMock{1}, nil
}

func TestValidateSession(t *testing.T) {
	const availableSessionID = "123"
	tests := []struct {
		name          string
		sessionID     string
		expectedErr   error
		expectedCalls int
	}{
		{"valid session ID", availableSessionID, nil, 1},
		{"invalid session ID", Modify(availableSessionID), service.ErrSessionNotFound, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &GetSessionMock{AvailableSessionID: availableSessionID}
			service := service.NewService(mock)

			err := service.ValidateSession(t.Context(), tt.sessionID)
			AssertErrorsEqual(t, tt.expectedErr, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}

func TestCreateSession(t *testing.T) {
	const availableUserID = 1
	tests := []struct {
		name          string
		userID        int64
		expectedErr   error
		expectedCalls int
	}{
		{"valid user ID", availableUserID, nil, 1},
		{"invalid user ID", availableUserID + 1, service.ErrSessionCreationFailed, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &CreateSessionMock{AvailableUserID: availableUserID}
			service := service.NewService(mock)

			_, err := service.CreateSession(t.Context(), tt.userID)
			AssertErrorsEqual(t, tt.expectedErr, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	const availableSessionID = "123"
	tests := []struct {
		name          string
		sessionID     string
		expectedErr   error
		expectedCalls int
	}{
		{"valid session ID", availableSessionID, nil, 1},
		{"invalid session ID", Modify(availableSessionID), service.ErrSessionNotFound, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &DeleteSessionMock{AvailableSessionID: availableSessionID}
			service := service.NewService(mock)

			err := service.DeleteSession(t.Context(), tt.sessionID)
			AssertErrorsEqual(t, tt.expectedErr, err)

			AssertCallCountsEqual(t, tt.expectedCalls, mock.Calls)
		})
	}
}
