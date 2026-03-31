package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/testutil"
)

func TestHandleLogin(t *testing.T) {
	user := testutil.RegisterUserParams{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password",
		IsAdmin:  true,
	}

	queries, _ := testutil.SetupTestDB(t)
	h := New(queries)

	testutil.SeedUser(t, queries, user)

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
			map[string]string{"username": user.Username},
			http.StatusUnauthorized,
		},
		{
			"invalid partial email form input",
			map[string]string{"username": user.Email},
			http.StatusUnauthorized,
		},
		{
			"invalid partial password form input",
			map[string]string{"password": user.Password},
			http.StatusUnauthorized,
		},
		{
			"valid username and password",
			map[string]string{"username": user.Username, "password": user.Password},
			http.StatusOK,
		},
		{
			"valid email and password",
			map[string]string{"username": user.Email, "password": user.Password},
			http.StatusOK,
		},
		{
			"invalid username, valid password",
			map[string]string{"username": fmt.Sprintf("not-%s", user.Username), "password": user.Password},
			http.StatusUnauthorized,
		},
		{
			"invalid email, valid password",
			map[string]string{"username": fmt.Sprintf("not-%s", user.Email), "password": user.Password},
			http.StatusUnauthorized,
		},
		{
			"valid username, invalid password",
			map[string]string{"username": user.Username, "password": fmt.Sprintf("not-%s", user.Password)},
			http.StatusUnauthorized,
		},
		{
			"valid email, invalid password",
			map[string]string{"username": user.Email, "password": fmt.Sprintf("not-%s", user.Password)},
			http.StatusUnauthorized,
		},
	}

	for _, tt := range requestTests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{"username": {user.Username}, "password": {user.Password}}.Encode()))
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
