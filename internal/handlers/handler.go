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
