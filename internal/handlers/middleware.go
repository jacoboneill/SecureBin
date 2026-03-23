package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/db"
)

type contextKey string

const UserIDContextKey contextKey = "userID"

func (h *Handler) htmx(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "invalid request type")
			return
		}
	}
}

func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		getSession := func(r *http.Request) (db.Session, error) {
			cookie, err := r.Cookie("session")
			if err != nil {
				return db.Session{}, err
			}
			return h.queries.GetSession(r.Context(), cookie.Value)
		}

		session, err := getSession(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
