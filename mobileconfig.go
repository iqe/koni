package main

import (
	"strings"

	"github.com/gofrs/uuid"
	macaron "gopkg.in/macaron.v1"
)

func mobileconfig(ctx *macaron.Context) {
	emailaddress := ctx.Req.FormValue("emailaddress")
	accountName := ctx.Req.FormValue("cn")
	password := ctx.Req.FormValue("password")

	payloadUUID1, _ := uuid.NewV4()
	payloadUUID2, _ := uuid.NewV4()

	data := map[string]interface{}{
		"emailaddress":         emailaddress,
		"account_name":         accountName,
		"smtp_server":          smtpServer,
		"imap_server":          imapServer,
		"payload_organization": provider,
		"payload_identifier":   createPayloadIdentifier(emailaddress, provider),
		"payload_uuid1":        payloadUUID1.String(),
		"payload_uuid2":        payloadUUID2.String(),
	}

	if password != "" {
		data["password"] = password
	}

	ctx.Render.HTML(200, "mobileconfig", data)
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
