package handlers

import (
	// "log/slog"
	"net/http"
)

func (h *Handler) PageFeed(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
