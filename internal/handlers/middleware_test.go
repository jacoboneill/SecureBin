package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTMXMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		isHTMX         bool
		expectedStatus int
	}{
		{"is htmx request", true, http.StatusOK},
		{"is not htmx request", false, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.isHTMX {
				r.Header.Set("HX-Request", "true")
			}

			w := httptest.NewRecorder()

			h := &Handler{}

			h.htmx(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
