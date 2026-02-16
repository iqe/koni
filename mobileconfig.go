package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/gofrs/uuid"
)

func sanitizeForHeader(s string) string {
	replacer := strings.NewReplacer("\"", "", "\r", "", "\n", "")
	return replacer.Replace(s)
}

func mobileconfigHandler(config koniConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailaddress := r.URL.Query().Get("emailaddress")
		if !validateEmail(emailaddress) {
			http.Error(w, "Invalid email address", http.StatusBadRequest)
			return
		}

		payloadUUID1, err := uuid.NewV4()
		if err != nil {
			log.Printf("koni: Failed to generate UUID: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		payloadUUID2, err := uuid.NewV4()
		if err != nil {
			log.Printf("koni: Failed to generate UUID: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		data := pongo2.Context{
			"emailaddress":         emailaddress,
			"account_name":         emailaddress,
			"smtp_server":          config.smtpServer,
			"imap_server":          config.imapServer,
			"payload_organization": config.provider,
			"payload_identifier":   createPayloadIdentifier(emailaddress, config.provider),
			"payload_uuid1":        payloadUUID1.String(),
			"payload_uuid2":        payloadUUID2.String(),
		}

		w.Header().Set("Content-Disposition", "attachment; filename=\""+sanitizeForHeader(emailaddress)+".mobileconfig\"")
		renderTemplate(w, "mobileconfig", http.StatusOK, data)
	}
}

func createPayloadIdentifier(emailaddress string, provider string) string {
	a := strings.Replace(emailaddress, "@", ".", -1)
	b := strings.Split(a, ".")
	b = append(b, "mobileconfig")
	b = append(b, provider)

	// reverse it
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	// webflow.mobileconfig.com.domain.user
	return strings.Join(b, ".")
}
