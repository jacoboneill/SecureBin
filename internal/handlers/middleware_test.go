package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/testutil"
)

func goodNext(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
	h := New(queries)

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

		h.auth(goodNext)(w, r)
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
