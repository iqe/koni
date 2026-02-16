package main

import (
	"regexp"
	"strings"

	macaron "gopkg.in/macaron.v1"
)

var emailRegexp = regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)

func autoconfigHandler(config koniConfig) macaron.Handler {
	return func(ctx *macaron.Context) {
		emailaddress := ctx.Req.URL.Query().Get("emailaddress")
		if !validateEmail(emailaddress) {
			ctx.Error(400, "Invalid email address")
			return
		}
		user, domain := splitEmail(emailaddress)

		data := map[string]interface{}{
			"provider":     config.provider,
			"domain":       domain,
			"emailaddress": emailaddress,
			"shortname":    user,
			"smtp_server":  config.smtpServer,
			"imap_server":  config.imapServer,
			"pop_server":   config.popServer,
		}

		ctx.Render.HTML(200, "autoconfig", data)
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
