package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-macaron/pongo2"
	"gopkg.in/macaron.v1"
)

var (
	configFileFlag = flag.String("c", "koni.conf", "Path to configuration file")
)

const (
	defaultListen   = "127.0.0.1:4443"
	defaultURL      = "https://acme-staging.api.letsencrypt.org/directory"
	defaultCertsDir = "."
)

func main() {
	flag.Parse()
	configFile := *configFileFlag
	if _, err := os.Stat(configFile); err != nil {
		log.Fatalf("Error: Cannot open config: %s", err)
	}


	// Remove date + time from logging output (systemd adds those for us)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	config := loadConfigFile(configFile)

	m := macaron.New()
	m.Use(apacheLogHandler())
	m.Use(macaron.Recovery())

	m.Use(pongo2.Pongoer(pongo2.Options{
		Directory:       "templates",
		Extensions:      []string{".xml.j2"},
		HTMLContentType: "text/xml",
	}))

	// Mozilla autoconfig
	m.Get("/autoconfig/config-v1.1.xml", autoconfigHandler(config))

	// Microsoft autodiscover v1
	m.Route("/autodiscover/autodiscover.xml", "GET, POST", autodiscoverHandler(config)) // GET support only for debugging
	m.Route("/Autodiscover/Autodiscover.xml", "GET, POST", autodiscoverHandler(config))

	// Let's Encrypt autocert via tls-alpn-01 challenge
	// See https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-01

	manager := buildAutocertManager(config.url, config.email, config.certsDir)

	s := &http.Server{
		Addr:         config.listen,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second, // Enough time to handle Let's Encrypt challenge on first request for any domain
		IdleTimeout:  120 * time.Second,
		Handler:      m,
		TLSConfig:    &tls.Config{GetCertificate: manager.GetCertificate},
	}

	log.Printf("Let's Encrypt URL: %s\n", config.url)
	log.Printf("Certificate cache directory: %s\n", config.certsDir)

	log.Printf("SMTP server: %s\n", config.smtpServer)
	log.Printf("IMAP server: %s\n", config.imapServer)
	log.Printf("POP3 server: %s\n", config.popServer)

	log.Printf("Starting HTTPS server on %s\n", config.listen)
	log.Println(s.ListenAndServeTLS("", ""))
}
