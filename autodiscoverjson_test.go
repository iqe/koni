package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAutodiscoverjsonHandler(t *testing.T) {
	handler := autodiscoverjsonHandler()

	t.Run("valid protocol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/autodiscover/autodiscover.json?Protocol=AutodiscoverV1&Email=user@example.com", nil)
		req.Host = "autodiscover.example.com"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}

		var resp autodiscoverjsonRedirect
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Protocol != "AutodiscoverV1" {
			t.Errorf("Protocol = %q, want %q", resp.Protocol, "AutodiscoverV1")
		}
		if resp.Url != "https://autodiscover.example.com/Autodiscover/Autodiscover.xml" {
			t.Errorf("Url = %q, want %q", resp.Url, "https://autodiscover.example.com/Autodiscover/Autodiscover.xml")
		}
	})

	t.Run("case insensitive protocol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/autodiscover/autodiscover.json?Protocol=autodiscoverv1", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("invalid protocol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/autodiscover/autodiscover.json?Protocol=invalid", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}

		var resp autodiscoverjsonError
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.ErrorCode != "InvalidProtocol" {
			t.Errorf("ErrorCode = %q, want %q", resp.ErrorCode, "InvalidProtocol")
		}
	})

	t.Run("missing protocol", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/autodiscover/autodiscover.json", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rr.Code, http.StatusBadRequest)
		}
	})
}
