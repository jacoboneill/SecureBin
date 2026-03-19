package handlers

import (
	"log/slog"
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/templates"
)

func (h *Handler) PageLogin(w http.ResponseWriter, r *http.Request) {
	if err := templates.Login("").Render(r.Context(), w); err != nil {
		slog.Error("template failed to load", "err", err)
	}
}
