package handler

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/jacoboneill/SecureBin/internal/contextkey"
	"github.com/jacoboneill/SecureBin/internal/db"
	"github.com/jacoboneill/SecureBin/internal/service"
	"github.com/jacoboneill/SecureBin/static"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

//	type Handler struct {
//		queries *db.Queries
//	}
//
//	func New(queries *db.Queries) *Handler {
//		return &Handler{
//			queries: queries,
//		}
//	}
func (h *Handler) NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Pages
	// mux.HandleFunc("GET /", h.PageFeed)
	mux.HandleFunc("GET /login", h.PageLogin)
	// mux.HandleFunc("GET /admin/register", h.auth(h.admin(h.PageRegister)))

	// Actions
	mux.HandleFunc("POST /login", h.htmx(h.HandleLogin))
	// mux.HandleFunc("POST /logout", h.htmx(h.auth(h.HandleLogout)))
	// mux.HandleFunc("POST /admin/register", h.htmx(h.auth(h.admin(h.HandleRegister))))

	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Files))))

	return mux
}

func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request, component templ.Component, status int) {
	ctx := r.Context()
	errorHelper := func(msg string, err error) {
		slog.Error(msg, "err", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}

	if _, ok := ctx.Value(contextkey.UserCtxKey).(*db.User); !ok {
		if cookie, err := r.Cookie("session"); err == nil {
			sessionID := cookie.Value
			user, err := h.service.GetUserFromSession(ctx, sessionID)
			if err == nil {
				ctx = context.WithValue(ctx, contextkey.UserCtxKey, user)
				ctx = context.WithValue(ctx, contextkey.SessionIDCtxKey, sessionID)
			}
		}
	}

	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		errorHelper("template failed to render", err)
		return
	}
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		errorHelper("buffer failed to write to HTTP response writer", err)
		return
	}
}
