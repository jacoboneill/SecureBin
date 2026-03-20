package handlers

import (
	"bytes"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jacoboneill/SecureBin/internal/templates"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, component templ.Component, statusCode int) {
	var buf bytes.Buffer
	if err := component.Render(r.Context(), &buf); err != nil {
		slog.Error("template failed to render", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	w.WriteHeader(statusCode)
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("buffer failed to write to HTTP response writer", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
}

func (h *Handler) PageLogin(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Return session key and redirect to "/"
	h.RenderTemplate(w, r, templates.LoginCallback("success!"), http.StatusOK)
}
