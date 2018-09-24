package main

import (
	"strings"

	macaron "gopkg.in/macaron.v1"
)

func autoconfig(ctx *macaron.Context) {
	emailaddress := ctx.Params("emailaddress")
	_, domain := splitEmail(emailaddress)

	data := map[string]interface{}{
		"provider":     provider,
		"domain":       domain,
		"emailaddress": emailaddress,
		"smtp_server":  smtpServer,
		"imap_server":  imapServer,
		"pop_server":   popServer,
	}

	ctx.Render.HTML(200, "autoconfig", data)
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
