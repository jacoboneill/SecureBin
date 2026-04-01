package handlers

import (
	"log/slog"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/contextkeys"
	"github.com/jacoboneill/SecureBin/internal/db"
)

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	httpInternalServerError := func() {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	sessionID, ok := ctx.Value(contextkeys.SessionIDCtxKey).(string)
	if !ok {
		slog.Error("sessionID not found in context")
		httpInternalServerError()
		return
	}

	if err := h.queries.DeleteSession(ctx, sessionID); err != nil {
		slog.Error("failed to delete session", "err", err)
		httpInternalServerError()
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	user, _ := ctx.Value(contextkeys.UserCtxKey).(*db.User)
	if user != nil {
		slog.Info("user logged out", "username", user.Username, "sessionID", sessionID)
	} else {
		slog.Info("user logged out", "sessionID", sessionID)
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
