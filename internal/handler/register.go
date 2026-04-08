package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/service"
	"github.com/jacoboneill/SecureBin/internal/template"
)

func (h *Handler) PageRegister(w http.ResponseWriter, r *http.Request) {
	h.RenderTemplate(w, r, template.Register(), http.StatusOK)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		username = r.FormValue("username")
		password = r.FormValue("password")
		isAdmin  = r.FormValue("isAdmin")
	)

	renderErr := func(msg string, status int, slogArgs ...any) {
		h.slog.Warn(msg, slogArgs...)
		h.RenderTemplate(w, r, template.RegisterCallback(msg, true), status)
	}

	if email == "" {
		renderErr("Email is invalid", http.StatusBadRequest, "email", email)
		return
	}
	if username == "" {
		renderErr("Username is invalid", http.StatusBadRequest, "username", username)
		return
	}
	if password == "" {
		renderErr("Password is invalid", http.StatusBadRequest, "password", password)
		return
	}
	if isAdmin != "on" && isAdmin != "" {
		renderErr("Admin checkbox is invalid", http.StatusBadRequest, "isAdmin", isAdmin)
		return
	}

	user, err := h.service.AddUser(r.Context(), username, email, password, isAdmin == "on")
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			renderErr("User already exists", http.StatusConflict, "username", username)
		} else {
			renderErr("something went wrong", http.StatusInternalServerError, "err", err)
		}
		return
	}

	h.slog.Info("new user added", "user", user)
	h.RenderTemplate(w, r, template.RegisterCallback(fmt.Sprintf("User %s registered successfully", user.Email), false), http.StatusOK)
}
