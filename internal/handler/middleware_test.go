package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
	"github.com/jacoboneill/SecureBin/internal/testutil"
)

func goodNext(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestLog(t *testing.T) {
	var buf bytes.Buffer
	h := &Handler{slog: slog.New(slog.NewJSONHandler(&buf, nil))}

	tests := []struct {
		name   string
		method string
		path   string
	}{{"valid request", http.MethodGet, "/"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			h.log(goodNext)(w, r)

			type logEntry struct {
				Message string `json:"msg"`
				Method  string `json:"method"`
				Path    string `json:"path"`
			}
			var entry logEntry
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("log output not valid JSON: %v\noutput: %s", err, buf.String())
			}
			if entry.Message != "http request" {
				t.Errorf("expected \"http request\", got %q", entry.Message)
			}
			if entry.Method != tt.method {
				t.Errorf("expected %q, got %q", tt.method, entry.Method)
			}
			if entry.Path != tt.path {
				t.Errorf("expected %q, got %q", tt.path, entry.Path)
			}
		})
	}
}

func TestHTMXMiddleware(t *testing.T) {
	h := &Handler{}

	tests := []struct {
		name           string
		isHTMX         bool
		expectedStatus int
	}{
		{"is htmx request", true, http.StatusOK},
		{"is not htmx request", false, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.isHTMX {
				r.Header.Set("HX-Request", "true")
			}

			w := httptest.NewRecorder()

			h.htmx(goodNext)(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func assertLoginRedirect(t testing.TB, w *httptest.ResponseRecorder) {
	t.Helper()

	resp := w.Result()
	statusCode := resp.StatusCode
	location := resp.Header.Get("Location")

	if statusCode != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, statusCode)
	}

	if location != "/login" {
		t.Errorf("expected redirect to /login, got %s", location)
	}
}

func TestAuthMiddleware(t *testing.T) {
	queries, _ := testutil.SetupTestDB(t)
	h := NewHandler(service.NewService(queries))

	user := testutil.SeedUser(t, queries, testutil.RegisterUserParams{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password",
		IsAdmin:  true,
	})

	t.Run("test no cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		h.auth(goodNext)(w, r)

		assertLoginRedirect(t, w)
	})

	t.Run("test invalid cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: "123456"})

		h.auth(goodNext)(w, r)

		assertLoginRedirect(t, w)
	})

	t.Run("test valid cookie", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "session", Value: user.SessionID})

		h.auth(func(w http.ResponseWriter, r *http.Request) {
			capturedUser, ok := r.Context().Value(contextkey.UserCtxKey).(*db.User)
			if !ok {
				t.Errorf("expected user in context")
			}
			if capturedUser.ID != user.ID {
				t.Errorf("expected userID %d in context, got %d", user.ID, capturedUser.ID)
			}
			capturedSessionID, ok := r.Context().Value(contextkey.SessionIDCtxKey).(string)
			if !ok {
				t.Errorf("expected session id in context")
			}
			if capturedSessionID != user.SessionID {
				t.Errorf("expected sessionID %q in context, got %q", user.SessionID, capturedSessionID)
			}
			w.WriteHeader(http.StatusOK)
		})(w, r)

		resp := w.Result()
		statusCode := resp.StatusCode
		location := resp.Header.Get("Location")

		if statusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, statusCode)
		}

		if location != "" {
			t.Errorf("did not expect redirect, got redirected to %s", location)
		}
	})
}

func TestAdmin(t *testing.T) {
	queries, _ := testutil.SetupTestDB(t)
	h := NewHandler(service.NewService(queries))

	admin := testutil.SeedUser(t, queries, testutil.RegisterUserParams{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password",
		IsAdmin:  true,
	})

	nonAdmin := testutil.SeedUser(t, queries, testutil.RegisterUserParams{
		Username: "non_admin",
		Email:    "non_admin@example.com",
		Password: "password",
		IsAdmin:  false,
	})

	tests := []struct {
		name           string
		user           testutil.User
		expectedStatus int
	}{
		{
			"test is admin",
			admin,
			http.StatusOK,
		},
		{
			"test is not admin",
			nonAdmin,
			http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.AddCookie(&http.Cookie{Name: "session", Value: tt.user.SessionID})

			var capturedUser *db.User
			h.auth(h.admin(func(w http.ResponseWriter, r *http.Request) {
				capturedUser, _ = r.Context().Value(contextkey.UserCtxKey).(*db.User)
				w.WriteHeader(http.StatusOK)
			}))(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				if capturedUser == nil {
					t.Fatal("expected user in context, got nil")
				}
				if !capturedUser.IsAdmin {
					t.Error("expected user to be admin")
				}
			}
		})
	}
}
