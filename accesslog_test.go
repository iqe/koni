package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"192.0.2.1:1234", "192.0.2.1"},
		{"[2001:db8::1]:1234", "2001:db8::1"},
		{"192.0.2.1", "192.0.2.1"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := clientIP(tt.input); got != tt.want {
				t.Errorf("clientIP(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAccessLogHandler(t *testing.T) {
	var buf bytes.Buffer
	defer slog.SetDefault(slog.Default())
	slog.SetDefault(slog.New(newLogHandler(&buf, slog.LevelInfo)))

	handler := accessLogHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest("GET", "/path?x=1", nil)
	req.RemoteAddr = "[2001:db8::1]:4711"
	handler.ServeHTTP(httptest.NewRecorder(), req)

	want := regexp.MustCompile(`^level=INFO msg=request client=2001:db8::1 host=example.com method=GET uri="/path\?x=1" proto=HTTP/1.1 status=418 bytes=5 duration=\S+\n$`)
	if !want.MatchString(buf.String()) {
		t.Errorf("unexpected log line: %q", buf.String())
	}
}
