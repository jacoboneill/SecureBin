package handler

import (
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/template"
)

func (h *Handler) PageFeed(w http.ResponseWriter, r *http.Request) {
	// HACK: Implement Feed page
	h.RenderTemplate(w, r, template.Feed(), http.StatusOK)
}
