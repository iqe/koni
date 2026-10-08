package main

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
)

func debugLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			slog.Error("Failed to dump request", "error", err)
		} else {
			slog.Debug("Request dump", "request", string(dump))
		}

		next.ServeHTTP(w, r)
	})
}
