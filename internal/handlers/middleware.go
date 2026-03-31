package handlers

import (
	"context"
	"database/sql"
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

		ctx := context.WithValue(r.Context(), userIDCtxKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpError := func() {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}

		ctx := r.Context()
		userID, ok := ctx.Value(userIDCtxKey).(int64)
		if !ok {
			slog.Error("failed to get user ID from context")
			httpError()
			return
		}

		user, err := h.queries.GetUser(ctx, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				slog.Warn("user not found", "userID", userID)
			} else {
				slog.Warn("query failed", "query", "GetUser", "err", err)
				httpError()
			}
			return
		}

		if !user.IsAdmin {
			slog.Warn("unauthorized user attempted to access admin page", "user", user.Username)
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(ctx, isAdminCtxKey, true)))
	}
}
