package main

import (
	"encoding/xml"

	macaron "gopkg.in/macaron.v1"
)

// Autodiscover request:
// <?xml version="1.0" encoding="utf-8"?>
// <Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
//   <Request>
//     <EMailAddress>User@domain.com</EMailAddress>
//     <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
//   </Request>
// </Autodiscover>

type Autodiscover struct {
	Request Request
}

type Request struct {
	EMailAddress             string
	AcceptableResponseSchema string
}

func autodiscoverHandler(config koniConfig) macaron.Handler {
	return func(ctx *macaron.Context) {
		b, err := ctx.Req.Body().Bytes()
		if err != nil {
			ctx.Error(400, "Invalid autodiscover request")
			return
		}
		var requestXML Autodiscover
		xml.Unmarshal(b, &requestXML)

		emailaddress := requestXML.Request.EMailAddress

		data := map[string]interface{}{
			"emailaddress": emailaddress,
			"smtp_server":  config.smtpServer,
			"imap_server":  config.imapServer,
			"pop_server":   config.popServer,
		}

		ctx.Render.HTML(200, "autodiscover", data)
	}
}
