package main

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

// Request categories for the access log. Everything that is not one of koni's
// endpoints (mostly bots scanning for vulnerabilities) is logged as noise.
const (
	categoryAutoconfig   = "autoconfig"
	categoryAutodiscover = "autodiscover"
	categoryMobileconfig = "mobileconfig"
	categoryACME         = "acme"
	categoryNoise        = "noise"
)

var pathCategories = map[string]string{
	"/mail/config-v1.1.xml":                        categoryAutoconfig,
	"/.well-known/autoconfig/mail/config-v1.1.xml": categoryAutoconfig,
	"/autodiscover/autodiscover.xml":               categoryAutodiscover,
	"/Autodiscover/Autodiscover.xml":               categoryAutodiscover,
	"/autodiscover/autodiscover.json":              categoryAutodiscover,
	"/mobileconfig.xml":                            categoryMobileconfig,
}

func requestCategory(path string) string {
	if category, ok := pathCategories[path]; ok {
		return category
	}
	if strings.HasPrefix(path, "/.well-known/acme-challenge/") {
		return categoryACME
	}
	return categoryNoise
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// Unwrap allows http.ResponseController to reach the underlying ResponseWriter
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func accessLogHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		startTime := time.Now()
		next.ServeHTTP(rec, r)
		elapsed := time.Since(startTime)

		slog.LogAttrs(r.Context(), slog.LevelInfo, "request",
			slog.String("category", requestCategory(r.URL.Path)),
			slog.String("client", clientIP(r.RemoteAddr)),
			slog.String("host", r.Host),
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("proto", r.Proto),
			slog.Int("status", rec.status),
			slog.Int("bytes", rec.size),
			slog.Duration("duration", elapsed),
		)
	})
}
