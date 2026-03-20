package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestHandleLogin(t *testing.T) {
	var (
		username = "admin"
		email    = "admin@example.com"
		password = "password"
		isAdmin  = true
	)

	queries, _ := testutil.SetupTestDB(t)
	h := New(queries)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := h.queries.RegisterUser(context.Background(), db.RegisterUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		IsAdmin:      isAdmin,
	}); err != nil {
		t.Fatal(err)
	}

	requestTests := []struct {
		name           string
		isHTMX         bool
		expectedStatus int
	}{
		{"htmx request returns partial", true, http.StatusOK},
		{"non-htmx request returns error", false, http.StatusBadRequest},
	}

	formTests := []struct {
		name           string
		form           map[string]string
		expectedStatus int
	}{
		{
			"invalid blank form input",
			map[string]string{},
			http.StatusUnauthorized,
		},
		{
			"invalid partial username form input",
			map[string]string{"username": username},
			http.StatusUnauthorized,
		},
		{
			"invalid partial email form input",
			map[string]string{"username": email},
			http.StatusUnauthorized,
		},
		{
			"invalid partial password form input",
			map[string]string{"password": password},
			http.StatusUnauthorized,
		},
		{
			"valid username and password",
			map[string]string{"username": username, "password": password},
			http.StatusOK,
		},
		{
			"valid email and password",
			map[string]string{"username": email, "password": password},
			http.StatusOK,
		},
		{
			"invalid username, valid password",
			map[string]string{"username": fmt.Sprintf("not-%s", username), "password": password},
			http.StatusUnauthorized,
		},
		{
			"invalid email, valid password",
			map[string]string{"username": fmt.Sprintf("not-%s", email), "password": password},
			http.StatusUnauthorized,
		},
		{
			"valid username, invalid password",
			map[string]string{"username": username, "password": fmt.Sprintf("not-%s", password)},
			http.StatusUnauthorized,
		},
		{
			"valid email, invalid password",
			map[string]string{"username": email, "password": fmt.Sprintf("not-%s", password)},
			http.StatusUnauthorized,
		},
	}

	for _, tt := range requestTests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{"username": {username}, "password": {password}}.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.isHTMX {
				r.Header.Set("HX-Request", "true")
			}

			w := httptest.NewRecorder()
			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}

	for _, tt := range formTests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			for k, val := range tt.form {
				form.Set(k, val)
			}

			r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("HX-Request", "true")

			w := httptest.NewRecorder()
			h.HandleLogin(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
