package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) htmx(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HX-Request") == "true" {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "invalid request type")
			return
		}
	}
}
