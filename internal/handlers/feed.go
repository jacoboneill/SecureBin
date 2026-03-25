package handlers

import (
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/templates"
)

func (h *Handler) PageFeed(w http.ResponseWriter, r *http.Request) {
	// HACK: Implement Feed page
	h.RenderTemplate(w, r, templates.Feed(), http.StatusOK)
}
