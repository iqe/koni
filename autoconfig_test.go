package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@domain.com", true},
		{"user@domain", true},
		{"a@b", true},
		{"", false},
		{"user", false},
		{"@domain.com", false},
		{"user @domain.com", false},
		{"user@ domain.com", false},
		{"user@\tdomain.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			got := validateEmail(tt.email)
			if got != tt.valid {
				t.Errorf("validateEmail(%q) = %v, want %v", tt.email, got, tt.valid)
			}
		})
	}
}

func TestSplitEmail(t *testing.T) {
	tests := []struct {
		input    string
		wantUser string
		wantDom  string
	}{
		{"user@domain.com", "user", "domain.com"},
		{"user@", "user", ""},
		{"@domain.com", "", "domain.com"},
		{"nodomain", "nodomain", ""},
		{"", "", ""},
		{"a@b@c", "a", "b"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			user, domain := splitEmail(tt.input)
			if user != tt.wantUser || domain != tt.wantDom {
				t.Errorf("splitEmail(%q) = (%q, %q), want (%q, %q)", tt.input, user, domain, tt.wantUser, tt.wantDom)
			}
		})
	}
}

func TestAutoconfigHandler(t *testing.T) {
	templateSet = pongo2.NewSet("templates", pongo2.MustNewLocalFileSystemLoader("templates"))

	config := koniConfig{
		provider:   "testprovider",
		smtpServer: "smtp.example.com",
		imapServer: "imap.example.com",
		popServer:  "pop.example.com",
	}
	handler := autoconfigHandler(config)

	t.Run("valid email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mail/config-v1.1.xml?emailaddress=user@example.com", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/xml; charset=utf-8" {
			t.Errorf("Content-Type = %q, want %q", ct, "text/xml; charset=utf-8")
		}
		body := rr.Body.String()
		if !contains(body, "smtp.example.com") {
			t.Error("response body missing smtp server")
		}
		if !contains(body, "user@example.com") {
			t.Error("response body missing email address")
		}
	})

	t.Run("missing email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mail/config-v1.1.xml", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/mail/config-v1.1.xml?emailaddress=notanemail", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
