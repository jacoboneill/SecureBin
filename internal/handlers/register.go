package handlers

import (
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/templates"
)

func (h *Handler) PageRegister(w http.ResponseWriter, r *http.Request) {
	h.RenderTemplate(w, r, templates.Register(), http.StatusOK)
}
