package main

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/flosch/pongo2/v6"
)

var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)

func autoconfigHandler(config koniConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		emailaddress := r.URL.Query().Get("emailaddress")
		if !validateEmail(emailaddress) {
			http.Error(w, "Invalid email address", http.StatusBadRequest)
			return
		}
		user, domain := splitEmail(emailaddress)

		data := pongo2.Context{
			"provider":     config.provider,
			"domain":       domain,
			"emailaddress": emailaddress,
			"shortname":    user,
			"smtp_server":  config.smtpServer,
			"imap_server":  config.imapServer,
			"pop_server":   config.popServer,
		}

		renderTemplate(w, "autoconfig", http.StatusOK, data)
	}
}

func validateEmail(email string) bool {
	return emailRegexp.MatchString(email)
}

func splitEmail(emailaddress string) (string, string) {
	parts := strings.Split(emailaddress, "@")

	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	if len(parts) >= 1 {
		return parts[0], ""
	}

	return emailaddress, ""
}
