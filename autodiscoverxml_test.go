package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestAutodiscoverxmlHandler(t *testing.T) {
	templateSet = pongo2.NewSet("templates", pongo2.MustNewLocalFileSystemLoader("templates"))

	config := koniConfig{
		smtpServer: "smtp.example.com",
		imapServer: "imap.example.com",
		popServer:  "pop.example.com",
	}
	handler := autodiscoverxmlHandler(config)

	t.Run("valid request", func(t *testing.T) {
		body := `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
  <Request>
    <EMailAddress>user@example.com</EMailAddress>
    <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
  </Request>
</Autodiscover>`
		req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader(body))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
		if ct := rr.Header().Get("Content-Type"); ct != "text/xml; charset=utf-8" {
			t.Errorf("Content-Type = %q, want %q", ct, "text/xml; charset=utf-8")
		}
	})

	t.Run("invalid XML", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader("not xml"))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid email in XML", func(t *testing.T) {
		body := `<?xml version="1.0" encoding="utf-8"?>
<Autodiscover><Request><EMailAddress>notanemail</EMailAddress></Request></Autodiscover>`
		req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader(body))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader(""))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("oversized body", func(t *testing.T) {
		// 65KB+ body should be rejected by MaxBytesReader
		bigBody := strings.Repeat("x", 65*1024+1)
		req := httptest.NewRequest("POST", "/autodiscover/autodiscover.xml", strings.NewReader(bigBody))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}
