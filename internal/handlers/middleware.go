package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		session, err := h.queries.GetSession(r.Context(), cookie.Value)
		if err != nil {
			slog.Error("failed to retrieve session from database", "err", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, session.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
