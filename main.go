package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-macaron/pongo2"
	"gopkg.in/macaron.v1"
)

var (
	addrFlag       = flag.String("l", "127.0.0.1:4443", "IP and port to listen on (HTTPs only)")
	urlFlag        = flag.String("u", "https://acme-staging.api.letsencrypt.org/directory", "Let's Encrypt URL")
	certsDirFlag   = flag.String("c", ".", "Certificate cache directory where certs are stored")
	smtpServerFlag = flag.String("s", "smtp.webflow.de", "SMTP server")
	imapServerFlag = flag.String("i", "imap.webflow.de", "IMAP server")
	popServerFlag  = flag.String("p", "pop.webflow.de", "POP server")
)

var (
	email      = "administrator@webflow.de"
	provider   = "webflow"
	smtpServer string
	imapServer string
	popServer  string
)

func main() {
	flag.Parse()
	addrs := *addrFlag
	url := *urlFlag
	certsDir := *certsDirFlag
	smtpServer = *smtpServerFlag
	imapServer = *imapServerFlag
	popServer = *popServerFlag

	m := macaron.New()
	m.Use(apacheLogHandler)
	m.Use(macaron.Recovery())

	m.Use(pongo2.Pongoer(pongo2.Options{
		Directory:       "templates",
		Extensions:      []string{".xml.j2"},
		HTMLContentType: "text/xml",
	}))

	// Mozilla autoconfig
	m.Get("/autoconfig/config-v1.1.xml", autoconfig)

	// Microsoft autodiscover v1
	m.Route("/autodiscover/autodiscover.xml", "GET, POST", autodiscover) // GET support only for debugging
	m.Route("/Autodiscover/Autodiscover.xml", "GET, POST", autodiscover)

	// Let's Encrypt autocert via tls-alpn-01 challenge
	// See https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-01

	manager := buildAutocertManager(url, email, certsDir)

	s := &http.Server{
		Addr:         addrs,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second, // Enough time to handle Let's Encrypt challenge on first request for any domain
		IdleTimeout:  120 * time.Second,
		Handler:      m,
		TLSConfig:    &tls.Config{GetCertificate: manager.GetCertificate},
	}

	log.Printf("Let's Encrypt URL: %s\n", url)
	log.Printf("Certificate cache directory: %s\n", certsDir)

	log.Printf("SMTP server: %s\n", smtpServer)
	log.Printf("IMAP server: %s\n", imapServer)
	log.Printf("POP3 server: %s\n", popServer)

	log.Printf("Starting HTTPS server on %s\n", addrs)
	log.Println(s.ListenAndServeTLS("", ""))
}
