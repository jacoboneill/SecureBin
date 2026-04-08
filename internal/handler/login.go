package handler

import (
	"errors"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/service"
	"github.com/jacoboneill/SecureBin/internal/template"
)

const sessionCookieName = "session"

func (h *Handler) PageLogin(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		sessionID := cookie.Value
		if err := h.service.ValidateSession(r.Context(), sessionID); err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	h.RenderTemplate(w, r, template.Login(), http.StatusOK)
}

var (
	ErrUsernameNotInForm = errors.New("username not present in form")
	ErrPasswordNotInForm = errors.New("password not present in form")
)

func validateFormInput(username, password string) error {
	if username == "" {
		return ErrUsernameNotInForm
	}

	if password == "" {
		return ErrPasswordNotInForm
	}
	return nil
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var (
		ctx      = r.Context()
		username = r.FormValue("username")
		password = r.FormValue("password")
	)

	authError := func() {
		h.RenderTemplate(w, r, template.LoginCallback("Username, email, or password are invalid"), http.StatusUnauthorized)
	}

	serverError := func() {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	if err := validateFormInput(username, password); err != nil {
		h.slog.Warn("malformed form input", "err", err)
		authError()
		return
	}

	user, err := h.service.AuthenticateUser(ctx, username, password)
	if err != nil {
		h.slog.Warn("authentication failed", "err", err)
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrInvalidPassword) {
			authError()
		} else {
			serverError()
		}
		return
	}

	sessionID, err := h.service.CreateSession(ctx, user.ID)
	if err != nil {
		h.slog.Error("failed to create Session", "err", err)
		serverError()
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	h.slog.Info("successful login", "user", user.Username, "sessionID", sessionID)
}
