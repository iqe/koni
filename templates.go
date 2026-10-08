package main

import (
	"bytes"
	"embed"
	"encoding/xml"
	"log/slog"
	"net/http"
	"text/template"
)

//go:embed templates/*.xml.tmpl
var templateFS embed.FS

// templates holds the response templates built into the binary
var templates = template.Must(template.New("").Option("missingkey=error").ParseFS(templateFS, "templates/*.xml.tmpl"))

// templateData holds template values. All values are XML-escaped before rendering.
type templateData map[string]string

func renderTemplate(w http.ResponseWriter, name string, contentType string, status int, data templateData) {
	escaped := make(templateData, len(data))
	for k, v := range data {
		var b bytes.Buffer
		xml.EscapeText(&b, []byte(v)) // writes to a bytes.Buffer never fail
		escaped[k] = b.String()
	}

	// Render into a buffer first so that errors can still produce a 500 response
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name+".xml.tmpl", escaped); err != nil {
		slog.Error("Failed to render template", "template", name, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}
