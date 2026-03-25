package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"log/slog"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) PageLogin(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		if _, err := h.queries.GetSession(r.Context(), cookie.Value); err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	h.RenderTemplate(w, r, templates.Login(""), http.StatusOK)
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var (
		username = r.FormValue("username")
		password = r.FormValue("password")
	)

	httpError := func() {
		h.RenderTemplate(w, r, templates.LoginCallback("Username, email, or password are invalid"), http.StatusUnauthorized)
	}

	// Check for malformed form
	if username == "" || password == "" {
		slog.Warn("username and password not present in request", "form", r.Form)
		httpError()
		return
	}

	// Check user exists
	user, err := h.queries.GetUserByEmailOrUsername(r.Context(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Warn("user not found", "username", username)
			httpError()
		} else {
			slog.Warn("query failed", "query", "GetUserByEmailOrUsername", "err", err)
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}
		return
	}

	// Check password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		slog.Warn("username and password do not match", "username", username)
		httpError()
		return
	}

	// Create Session
	token := make([]byte, 32)
	rand.Read(token)
	sessionID := base64.URLEncoding.EncodeToString(token)

	session, err := h.queries.CreateSession(r.Context(), db.CreateSessionParams{ID: sessionID, UserID: user.ID})
	if err != nil {
		slog.Error("failed to add new session to database", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
	slog.Info("successful login", "user", user.Username, "sessionID", session.ID)
}
