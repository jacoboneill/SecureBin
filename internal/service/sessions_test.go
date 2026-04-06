package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
)

type QuerierMock struct {
	db.Querier
	Calls int
}

type GetSessionMock struct {
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

func AssertErrorsEqual(t testing.TB, expectedErr, capturedErr error) {
	t.Helper()
	if !errors.Is(capturedErr, expectedErr) {
		if expectedErr == nil {
			t.Errorf("expected no errors, got %v", capturedErr)
		} else {
			t.Errorf("expected error: %v, got %v", expectedErr, capturedErr)
		}
	}
}

func AssertCallCountsEqual(t testing.TB, expectedCallCount, capturedCallCount int) {
	t.Helper()
	if expectedCallCount != capturedCallCount {
		var suffix string
		if expectedCallCount > 1 {
			suffix = "s"
		}
		t.Errorf("expected %d %s, got %d", expectedCallCount, suffix, capturedCallCount)
	}
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
		{"invalid session ID", fmt.Sprintf("%s2", availableSessionID), service.ErrSessionNotFound, 1},
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
