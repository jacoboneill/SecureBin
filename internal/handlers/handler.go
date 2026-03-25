package handlers

import (
	"bytes"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
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
	mux.HandleFunc("GET /admin/register", h.auth(h.admin(h.PageRegister)))

	// Actions
	mux.HandleFunc("POST /login", h.htmx(h.HandleLogin))

	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Files))))

	return mux
}

func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, component templ.Component, statusCode int) {
	var buf bytes.Buffer
	if err := component.Render(r.Context(), &buf); err != nil {
		slog.Error("template failed to render", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	w.WriteHeader(statusCode)
	if _, err := buf.WriteTo(w); err != nil {
		slog.Error("buffer failed to write to HTTP response writer", "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
}
