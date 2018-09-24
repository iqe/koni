package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/go-macaron/pongo2"
	"github.com/pelletier/go-toml"
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

var (
	email      string
	provider   string
	smtpServer string
	imapServer string
	popServer  string
)

func main() {
	flag.Parse()
	configFile := *configFileFlag
	config, err := toml.LoadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to load config file %s: %s", configFile, err)
	}

	addr := config.GetDefault("listen", defaultListen).(string)

	url := config.GetDefault("letsencrypt.url", defaultURL).(string)
	certsDir := config.GetDefault("letsencrypt.certs_dir", defaultCertsDir).(string)
	email := config.Get("letsencrypt.email").(string)

	provider = config.Get("mail.provider_id").(string)
	imapServer = config.Get("mail.imap_server").(string)
	popServer = config.Get("mail.pop3_server").(string)
	smtpServer = config.Get("mail.smtp_server").(string)

	// Remove date + time from logging output (systemd adds those for us)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

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
		Addr:         addr,
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

	log.Printf("Starting HTTPS server on %s\n", addr)
	log.Println(s.ListenAndServeTLS("", ""))
}
