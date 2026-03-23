package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) PageFeed(w http.ResponseWriter, r *http.Request) {
	// HACK: Implement Feed page
	fmt.Fprint(w, "TODO: feed page")
}
