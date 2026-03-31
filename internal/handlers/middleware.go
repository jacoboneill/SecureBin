package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/contextkeys"
	"github.com/jacoboneill/SecureBin/internal/db"
)

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

		user, err := h.queries.GetUser(ctx, session.UserID)
		if err != nil {
			slog.Warn("user not found but session ID found", "err", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, contextkeys.UserCtxKey, &user)))
	}
}

func (h *Handler) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := ctx.Value(contextkeys.UserCtxKey).(*db.User)
		if !ok {
			slog.Error("user not found in context. please ensure to use auth middleware before admin")
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		if !user.IsAdmin {
			slog.Warn("unauthorized user attempted to access admin page", "user", user.Username)
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
