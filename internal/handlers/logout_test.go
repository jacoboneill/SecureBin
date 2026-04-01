package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/contextkeys"
	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/testutil"
)

type logoutRequestConfig struct {
	isHTMX     bool
	hasCookie  bool
	hasContext bool
}

func createLogoutRequest(t *testing.T, q *db.Queries, cfg logoutRequestConfig) (*http.Request, testutil.User) {
	t.Helper()

	r := httptest.NewRequest(http.MethodPost, "/logout", nil)
	if cfg.isHTMX {
		r.Header.Set("HX-Request", "true")
	}

	var user testutil.User
	if cfg.hasCookie || cfg.hasContext {
		user = testutil.SeedUser(t, q, testutil.RegisterUserParams{
			Username: "test",
			Email:    "test@example.com",
			Password: "password",
			IsAdmin:  false,
		})
	}

	if cfg.hasCookie {
		r.AddCookie(&http.Cookie{Name: "session", Value: user.SessionID})
	}

	if cfg.hasContext {
		dbUser := db.User{ID: user.ID, Username: user.Username, IsAdmin: user.IsAdmin}
		ctx := context.WithValue(r.Context(), contextkeys.UserCtxKey, &dbUser)
		ctx = context.WithValue(ctx, contextkeys.SessionIDCtxKey, user.SessionID)
		r = r.WithContext(ctx)
	}

	return r, user
}

func TestHandleLogoutHTMX(t *testing.T) {
	tests := []struct {
		name           string
		config         logoutRequestConfig
		expectedStatus int
	}{
		{
			name:           "valid htmx request",
			config:         logoutRequestConfig{isHTMX: true, hasCookie: true, hasContext: false},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-htmx request",
			config:         logoutRequestConfig{isHTMX: false, hasCookie: true, hasContext: false},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _ := testutil.SetupTestDB(t)
			h := New(q)

			r, _ := createLogoutRequest(t, q, tt.config)
			w := httptest.NewRecorder()
			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleLogoutAuth(t *testing.T) {
	tests := []struct {
		name           string
		config         logoutRequestConfig
		expectedStatus int
	}{
		{
			name:           "no cookie redirects to login",
			config:         logoutRequestConfig{isHTMX: true, hasCookie: false, hasContext: false},
			expectedStatus: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _ := testutil.SetupTestDB(t)
			h := New(q)

			r, _ := createLogoutRequest(t, q, tt.config)
			w := httptest.NewRecorder()
			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if loc := w.Header().Get("Location"); loc != "/login" {
				t.Errorf("expected redirect to /login, got %q", loc)
			}
		})
	}
}

func TestHandleLogoutContext(t *testing.T) {
	tests := []struct {
		name           string
		config         logoutRequestConfig
		expectedStatus int
		expectDeleted  bool
		expectRefresh  bool
	}{
		{
			name:           "with context deletes session",
			config:         logoutRequestConfig{isHTMX: false, hasCookie: false, hasContext: true},
			expectedStatus: http.StatusOK,
			expectDeleted:  true,
			expectRefresh:  true,
		},
		{
			name:           "with cookie and context deletes session",
			config:         logoutRequestConfig{isHTMX: false, hasCookie: true, hasContext: true},
			expectedStatus: http.StatusOK,
			expectDeleted:  true,
			expectRefresh:  true,
		},
		{
			name:           "no context or cookie returns 500",
			config:         logoutRequestConfig{isHTMX: false, hasCookie: false, hasContext: false},
			expectedStatus: http.StatusInternalServerError,
			expectDeleted:  false,
			expectRefresh:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _ := testutil.SetupTestDB(t)
			h := New(q)

			r, user := createLogoutRequest(t, q, tt.config)
			w := httptest.NewRecorder()
			h.HandleLogout(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectRefresh {
				if w.Header().Get("HX-Refresh") != "true" {
					t.Error("expected HX-Refresh header to be true")
				}
			}

			if tt.expectDeleted {
				if _, err := q.GetSession(t.Context(), user.SessionID); err == nil {
					t.Error("expected session to be deleted from database")
				} else if !errors.Is(err, sql.ErrNoRows) {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
