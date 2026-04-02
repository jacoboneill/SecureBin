package handler

import (
	"log/slog"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/contextkey"
)

func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	internalServerError := func() {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	sessionID, ok := ctx.Value(contextkey.SessionIDCtxKey).(string)
	if !ok {
		slog.Error("sessionID was not found in context")
		internalServerError()
		return
	}

	if err := h.service.DeleteSession(ctx, sessionID); err != nil {
		slog.Error("unable to delete session", "err", err)
		internalServerError()
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	if user, _ := h.service.GetUserFromContext(ctx); user != nil {
		slog.Info("user logged out", "username", user.Username, "sessionID", sessionID)
	} else {
		slog.Info("user logged out", "sessionID", sessionID)
	}

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}
