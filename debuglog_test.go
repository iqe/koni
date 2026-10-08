package main

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDebugLogHandler(t *testing.T) {
	var buf bytes.Buffer
	defer slog.SetDefault(slog.Default())
	slog.SetDefault(slog.New(newLogHandler(&buf, slog.LevelDebug)))

	called := false
	handler := debugLogHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader("<body/>"))
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if !called {
		t.Error("next handler was not called")
	}
	out := buf.String()
	if !strings.Contains(out, "level=DEBUG") || !strings.Contains(out, "POST /autodiscover/autodiscover.xml") || !strings.Contains(out, "<body/>") {
		t.Errorf("unexpected debug output: %q", out)
	}
}
