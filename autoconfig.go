package main

import (
	"strings"

	macaron "gopkg.in/macaron.v1"
)

func autoconfigHandler(config koniConfig) macaron.Handler {
	return func(ctx *macaron.Context) {
		emailaddress := ctx.Params("emailaddress")
		_, domain := splitEmail(emailaddress)

		data := map[string]interface{}{
			"provider":     config.provider,
			"domain":       domain,
			"emailaddress": emailaddress,
			"smtp_server":  config.smtpServer,
			"imap_server":  config.imapServer,
			"pop_server":   config.popServer,
		}

		ctx.Render.HTML(200, "autoconfig", data)
	}
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
