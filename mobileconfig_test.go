package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestSanitizeForHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"clean", "user@example.com", "user@example.com"},
		{"quotes", `user"@"example.com`, "user@example.com"},
		{"newlines", "user\r\n@example.com", "user@example.com"},
		{"control chars CR", "user\r@example.com", "user@example.com"},
		{"control chars LF", "user\n@example.com", "user@example.com"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeForHeader(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeForHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreatePayloadIdentifier(t *testing.T) {
	tests := []struct {
		email    string
		provider string
		want     string
	}{
		{"user@domain.com", "webflow", "webflow.mobileconfig.com.domain.user"},
		{"alice@sub.domain.org", "myprovider", "myprovider.mobileconfig.org.domain.sub.alice"},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := createPayloadIdentifier(tt.email, tt.provider)
			if got != tt.want {
				t.Errorf("createPayloadIdentifier(%q, %q) = %q, want %q", tt.email, tt.provider, got, tt.want)
			}
		})
	}
}

func TestMobileconfigHandler(t *testing.T) {
	templateSet = pongo2.NewSet("templates", pongo2.MustNewLocalFileSystemLoader("templates"))

	config := koniConfig{
		provider:   "testprovider",
		smtpServer: "smtp.example.com",
		imapServer: "imap.example.com",
		popServer:  "pop.example.com",
	}
	handler := mobileconfigHandler(config)

	t.Run("valid email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mobileconfig.xml?emailaddress=user@example.com", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		cd := rr.Header().Get("Content-Disposition")
		if cd == "" {
			t.Error("missing Content-Disposition header")
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mobileconfig.xml?emailaddress=bad", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}
