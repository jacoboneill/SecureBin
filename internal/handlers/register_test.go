package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jacoboneill/SecureBin/internal/testutil"
)

type Config struct {
	seedRegistrar    bool
	isRegistrarAdmin bool
	seedExistingUser *testutil.RegisterUserParams
}

func (c Config) SetupDatabase(t *testing.T) (*Handler, testutil.User) {
	t.Helper()

	queries, _ := testutil.SetupTestDB(t)
	h := New(queries)

	if c.seedExistingUser != nil {
		testutil.SeedUser(t, queries, *c.seedExistingUser)
	}

	if c.seedRegistrar {
		admin := testutil.SeedUser(t, queries, testutil.RegisterUserParams{
			Username: "admin",
			Email:    "admin@example.com",
			Password: "password",
			IsAdmin:  c.isRegistrarAdmin,
		})
		return h, admin
	}

	return h, testutil.User{}
}

func convertToReader(t *testing.T, form map[string]string) *strings.Reader {
	t.Helper()

	vals := url.Values{}
	for k, v := range form {
		vals.Add(k, v)
	}

	return strings.NewReader(vals.Encode())
}

type RegisterUserForm struct {
	testutil.RegisterUserParams
}

func (ruf RegisterUserForm) ConvertToReader(t *testing.T) io.Reader {
	t.Helper()

	form := map[string]string{
		"email":    ruf.Email,
		"username": ruf.Username,
		"password": ruf.Password,
	}

	if ruf.IsAdmin {
		form["isAdmin"] = "on"
	}

	return convertToReader(t, form)
}

func TestHandleRegisterHTMX(t *testing.T) {
	queries, _ := testutil.SetupTestDB(t)
	h := New(queries)
	admin := testutil.SeedUser(t, queries, testutil.RegisterUserParams{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password",
		IsAdmin:  true,
	})

	newUser := testutil.RegisterUserParams{
		Username: "test",
		Email:    "test@example.com",
		Password: "password",
		IsAdmin:  false,
	}

	tests := []struct {
		name           string
		isHTMX         bool
		expectedStatus int
	}{
		{
			name:           "htmx request succeeds",
			isHTMX:         true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-htmx request returns 400",
			isHTMX:         false,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/admin/register", RegisterUserForm{newUser}.ConvertToReader(t))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.isHTMX {
				r.Header.Set("HX-Request", "true")
			}
			r.AddCookie(&http.Cookie{Name: "session", Value: admin.SessionID})

			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleRegisterMalformedForm(t *testing.T) {
	config := Config{true, true, nil}

	tests := []struct {
		name           string
		form           map[string]string
		expectedStatus int
	}{
		{
			name:           "missing all fields",
			form:           map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing email",
			form:           map[string]string{"username": "test", "password": "password"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing username",
			form:           map[string]string{"email": "test@example.com", "password": "password"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing password",
			form:           map[string]string{"email": "test@example.com", "username": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "isAdmin checkbox on",
			form:           map[string]string{"email": "test@example.com", "username": "test", "password": "password", "isAdmin": "on"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "isAdmin checkbox omitted",
			form:           map[string]string{"email": "test@example.com", "username": "test", "password": "password"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "isAdmin checkbox invalid value",
			form:           map[string]string{"email": "test@example.com", "username": "test", "password": "password", "isAdmin": "yes"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, admin := config.SetupDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/admin/register", convertToReader(t, tt.form))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("HX-Request", "true")
			r.AddCookie(&http.Cookie{Name: "session", Value: admin.SessionID})

			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleRegisterAuthz(t *testing.T) {
	newUser := testutil.RegisterUserParams{
		Username: "test",
		Email:    "test@example.com",
		Password: "password",
		IsAdmin:  false,
	}

	tests := []struct {
		name               string
		newUser            testutil.RegisterUserParams
		setupConfig        Config
		expectedInDatabase bool
		expectedStatus     int
	}{
		{
			name:               "admin registers user",
			newUser:            newUser,
			setupConfig:        Config{true, true, nil},
			expectedInDatabase: true,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "non-admin gets 403",
			newUser:            newUser,
			setupConfig:        Config{true, false, nil},
			expectedInDatabase: false,
			expectedStatus:     http.StatusForbidden,
		},
		{
			name:               "no session gets redirect",
			newUser:            newUser,
			setupConfig:        Config{false, false, nil},
			expectedInDatabase: false,
			expectedStatus:     http.StatusSeeOther,
		},
		{
			name:               "duplicate user gets 409",
			newUser:            newUser,
			setupConfig:        Config{true, true, &newUser},
			expectedInDatabase: true,
			expectedStatus:     http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, admin := tt.setupConfig.SetupDatabase(t)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/admin/register", RegisterUserForm{tt.newUser}.ConvertToReader(t))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Set("HX-Request", "true")
			r.AddCookie(&http.Cookie{Name: "session", Value: admin.SessionID})

			h.NewRouter().ServeHTTP(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			_, err := h.queries.GetUserByEmailOrUsername(t.Context(), tt.newUser.Email)
			if tt.expectedInDatabase && err != nil {
				t.Errorf("expected user in database, got error: %v", err)
			}
			if !tt.expectedInDatabase && err == nil {
				t.Errorf("expected user not to be in database, but found one")
			}
		})
	}
}
