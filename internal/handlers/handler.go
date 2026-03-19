package handlers

import (
	"net/http"

	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/static"
)

type Handler struct {
	queries *db.Queries
}

func New(queries *db.Queries) *Handler {
	return &Handler{
		queries: queries,
	}
}

func (h *Handler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Pages
	mux.HandleFunc("GET /", h.PageFeed)
	mux.HandleFunc("GET /login", h.PageLogin)

	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Files))))

	return mux
}
