package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/service"
)

func (h *Handler) log(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.slog.Info("http request", "method", r.Method, "path", r.URL.Path)
		next(w, r)
	}
}

func (h *Handler) htmx(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			if _, err := fmt.Fprint(w, "invalid request type"); err != nil {
				h.slog.Error("failed to write to response writer", "err", err)
			}
			return
		}
	}
}

func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		httpRedirectLogin := func() {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		warnErrorHelper := func(msg, sessionID string, err, warningError error) {
			if errors.Is(err, warningError) {
				h.slog.Warn(msg, "err", err, "sessionID", sessionID)
			} else {
				h.slog.Error(msg, "err", err, "sessionID", sessionID)
			}
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			httpRedirectLogin()
			return
		}
		sessionID := cookie.Value

		if err := h.service.ValidateSession(ctx, sessionID); err != nil {
			warnErrorHelper("failed to get session from database", sessionID, err, service.ErrSessionNotFound)
			httpRedirectLogin()
			return
		}

		user, err := h.service.GetUserFromSession(ctx, sessionID)
		if err != nil {
			warnErrorHelper("failed to get user from database", sessionID, err, service.ErrUserNotFound)
			return
		}

		ctx = context.WithValue(ctx, contextkey.SessionIDCtxKey, sessionID)
		ctx = context.WithValue(ctx, contextkey.UserCtxKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (h *Handler) admin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		internalServerError := func() {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}

		forbidden := func() {
			http.Error(w, "forbidden", http.StatusForbidden)
		}

		ctx := r.Context()

		user, err := h.service.GetUserFromContext(ctx)
		if err != nil {
			h.slog.Error("failed to get user from context", "err", err)
			internalServerError()
			return
		}

		if !user.IsAdmin {
			h.slog.Warn("unauthorized user attempted to access admin page", "user", user.Username)
			forbidden()
			return
		}

		next.ServeHTTP(w, r)
	}
}
