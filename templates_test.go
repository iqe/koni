package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderTemplateEscapesXML(t *testing.T) {
	rr := httptest.NewRecorder()
	renderTemplate(rr, "autodiscover", contentTypeXML, http.StatusOK, templateData{
		"emailaddress": `a&b<c>"d'@example.com`,
		"smtp_server":  "smtp.example.com",
		"imap_server":  "imap.example.com",
		"pop_server":   "pop.example.com",
	})

	body := rr.Body.String()
	if strings.Contains(body, `a&b<c>`) {
		t.Errorf("value was not escaped: %s", body)
	}
	if !strings.Contains(body, "a&amp;b&lt;c&gt;&#34;d&#39;@example.com") {
		t.Errorf("escaped value not found in body: %s", body)
	}
}

func TestRenderTemplateMissingKey(t *testing.T) {
	rr := httptest.NewRecorder()
	renderTemplate(rr, "autodiscover", contentTypeXML, http.StatusOK, templateData{})

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestEmbeddedTemplates(t *testing.T) {
	for _, name := range []string{"autoconfig", "autodiscover", "mobileconfig"} {
		if templates.Lookup(name+".xml.tmpl") == nil {
			t.Errorf("template %s.xml.tmpl is not embedded", name)
		}
	}
}
