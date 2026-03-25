package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/templates"
	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func (h *Handler) PageRegister(w http.ResponseWriter, r *http.Request) {
	h.RenderTemplate(w, r, templates.Register(), http.StatusOK)
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Validate form
	var (
		formInputEmail    = r.FormValue("email")
		formInputUsername = r.FormValue("username")
		formInputPassword = r.FormValue("password")
		formInputIsAdmin  = r.FormValue("isAdmin")
	)

	// callback is an internal helper for specific callback functions. Do not use directly.
	callback := func(slogMessage, htmlMessage string, isError bool, statusCode int, slogArgs ...any) {
		slog.Warn(slogMessage, slogArgs...)
		h.RenderTemplate(w, r, templates.RegisterCallback(htmlMessage, isError), statusCode)
	}

	validationErrorCallback := func(msg string) {
		callback(strings.ToLower(msg), msg, true, http.StatusBadRequest, "email", formInputEmail, "username", formInputUsername, "password", formInputPassword, "isAdmin", formInputIsAdmin)
	}

	if formInputEmail == "" {
		validationErrorCallback("Email is invalid")
		return
	}
	if formInputUsername == "" {
		validationErrorCallback("Username is invalid")
		return
	}
	if formInputPassword == "" {
		validationErrorCallback("Password is invalid")
		return
	}
	if formInputIsAdmin != "on" && formInputIsAdmin != "" {
		validationErrorCallback("Admin checkbox is invalid")
		return
	}

	// Add user
	registrationErrorCallback := func(slogMessage string, err error) {
		callback(slogMessage, "something went wrong", true, http.StatusInternalServerError, "err", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(formInputPassword), bcrypt.DefaultCost)
	if err != nil {
		registrationErrorCallback("bcrypt failed to hash password", err)
	}

	user := db.RegisterUserParams{
		Username:     formInputUsername,
		Email:        formInputEmail,
		PasswordHash: string(passwordHash),
		IsAdmin:      formInputIsAdmin == "on",
	}
	if _, err := h.queries.RegisterUser(r.Context(), user); err != nil {
		var sqliteErr *sqlite.Error
		errors.As(err, &sqliteErr)
		if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			callback("registrar attempted to register user that already exists", "User already exists", true, http.StatusConflict, "registrar", r.Context().Value(userIDCtxKey), "registree", user.Email)
		} else {
			registrationErrorCallback("RegisterUser query failed", err)
		}
		return
	}

	slog.Info("new user added", "user", user)
	h.RenderTemplate(w, r, templates.RegisterCallback(fmt.Sprintf("User %s registered successfully", user.Email), false), http.StatusOK)
}
