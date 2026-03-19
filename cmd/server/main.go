package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/jacoboneill/SecureBin/internal/handlers"
)

func main() {
	h := handlers.New()

	const port = 8080

	slog.Info("server started", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), h.NewRouter()); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
