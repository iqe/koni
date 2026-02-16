package main

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"

	"github.com/flosch/pongo2/v6"
)

// Autodiscover request:
// <?xml version="1.0" encoding="utf-8"?>
// <Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
//   <Request>
//     <EMailAddress>User@domain.com</EMailAddress>
//     <AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
//   </Request>
// </Autodiscover>

type autodiscover struct {
	Request request
}

type request struct {
	EMailAddress             string
	AcceptableResponseSchema string
}

func autodiscoverxmlHandler(config koniConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)

		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("koni: Failed to read autodiscover body bytes: %v\n", err)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		var requestXML autodiscover
		if err := xml.Unmarshal(b, &requestXML); err != nil {
			log.Printf("koni: Failed to parse autodiscover XML: %v\n", err)
			http.Error(w, "Invalid XML", http.StatusBadRequest)
			return
		}

		emailaddress := requestXML.Request.EMailAddress
		if !validateEmail(emailaddress) {
			http.Error(w, "Invalid email address", http.StatusBadRequest)
			return
		}

		data := pongo2.Context{
			"emailaddress": emailaddress,
			"smtp_server":  config.smtpServer,
			"imap_server":  config.imapServer,
			"pop_server":   config.popServer,
		}

		renderTemplate(w, "autodiscover", http.StatusOK, data)
	}
}
