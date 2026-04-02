package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
	"github.com/jacoboneill/SecureBin/internal/template"
	"github.com/jacoboneill/SecureBin/internal/testutil"
)

func AssertInDoc(t testing.TB, doc *goquery.Document, selector, errorMessage string) {
	t.Helper()
	if doc.Find(selector).Length() == 0 {
		t.Error(errorMessage)
	}
}

func AssertNotInDoc(t testing.TB, doc *goquery.Document, selector, errorMessage string) {
	t.Helper()
	if doc.Find(selector).Length() != 0 {
		t.Error(errorMessage)
	}
}

type navItem struct {
	selector string
	label    string
}

func TestRenderTemplateNavigation(t *testing.T) {
	tests := []struct {
		name        string
		user        *testutil.RegisterUserParams
		useContext  bool
		expectedNav []navItem
	}{
		{
			name: "unauthenticated",
			user: nil,
			expectedNav: []navItem{
				{`a[href="/login"]`, "Login"},
			},
		},
		{
			name: "authenticated non-admin via context",
			user: &testutil.RegisterUserParams{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password",
				IsAdmin:  false,
			},
			useContext: true,
			expectedNav: []navItem{
				{`a[href="/p/new"]`, "New Paste"},
				{`a[href="/testuser"]`, "Account"},
				{`button[hx-post="/logout"]`, "Logout"},
			},
		},
		{
			name: "authenticated admin via context",
			user: &testutil.RegisterUserParams{
				Username: "admin",
				Email:    "admin@example.com",
				Password: "password",
				IsAdmin:  true,
			},
			useContext: true,
			expectedNav: []navItem{
				{`a[href="/p/new"]`, "New Paste"},
				{`a[href="/admin/register"]`, "Register New User"},
				{`a[href="/admin"]`, "Account"},
				{`button[hx-post="/logout"]`, "Logout"},
			},
		},
		{
			name: "authenticated non-admin via cookie fallback",
			user: &testutil.RegisterUserParams{
				Username: "cookieuser",
				Email:    "cookie@example.com",
				Password: "password",
				IsAdmin:  false,
			},
			useContext: false,
			expectedNav: []navItem{
				{`a[href="/p/new"]`, "New Paste"},
				{`a[href="/cookieuser"]`, "Account"},
				{`button[hx-post="/logout"]`, "Logout"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, _ := testutil.SetupTestDB(t)
			h := NewHandler(service.NewService(queries))

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.user != nil {
				seeded := testutil.SeedUser(t, queries, *tt.user)
				if tt.useContext {
					dbUser := db.User{ID: seeded.ID, Username: seeded.Username, IsAdmin: tt.user.IsAdmin}
					ctx := context.WithValue(r.Context(), contextkey.UserCtxKey, &dbUser)
					r = r.WithContext(ctx)
				} else {
					r.AddCookie(&http.Cookie{Name: "session", Value: seeded.SessionID})
				}
			}

			h.RenderTemplate(w, r, template.Base("test"), http.StatusOK)

			doc, err := goquery.NewDocumentFromReader(w.Body)
			if err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Verify expected nav items exist and are in order
			navItems := doc.Find("nav ul li")
			if navItems.Length() != len(tt.expectedNav) {
				t.Fatalf("expected %d nav items, got %d", len(tt.expectedNav), navItems.Length())
			}

			for i, expected := range tt.expectedNav {
				li := navItems.Eq(i)
				el := li.Find(expected.selector)
				if el.Length() == 0 {
					t.Errorf("nav item %d: expected %q (%s), not found", i, expected.label, expected.selector)
					continue
				}
				if got := el.Text(); got != expected.label {
					t.Errorf("nav item %d: expected label %q, got %q", i, expected.label, got)
				}
			}
		})
	}
}
