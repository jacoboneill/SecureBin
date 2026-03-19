package handlers

import (
	"net/http"

	"github.com/jacoboneill/SecureBin/static"
	// "github.com/jacoboneill/SecureBin/internal/db"
)

// type Handler struct {
// 	queries *db.Queries
// }

// func New(queries *db.Queries) *Handler {
// 	return &Handler{
// 		queries: queries,
// 	}
// }

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Pages
	mux.HandleFunc("GET /login", h.PageLogin)

	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Files))))

	return mux
}
