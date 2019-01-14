package main

import (
	"strings"

	"github.com/gofrs/uuid"
	macaron "gopkg.in/macaron.v1"
)

func mobileconfigHandler(config koniConfig) macaron.Handler {
	return func(ctx *macaron.Context) {
		emailaddress := ctx.Req.URL.Query().Get("emailaddress")

		payloadUUID1, _ := uuid.NewV4()
		payloadUUID2, _ := uuid.NewV4()

		data := map[string]interface{}{
			"emailaddress":         emailaddress,
			"account_name":         emailaddress,
			"smtp_server":          config.smtpServer,
			"imap_server":          config.imapServer,
			"payload_organization": config.provider,
			"payload_identifier":   createPayloadIdentifier(emailaddress, config.provider),
			"payload_uuid1":        payloadUUID1.String(),
			"payload_uuid2":        payloadUUID2.String(),
		}

		ctx.Resp.Header().Set("Content-Disposition", "attachment; filename=\"" + emailaddress + ".mobileconfig\"")
		ctx.Render.HTML(200, "mobileconfig", data)
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
