package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
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
	if _, err := os.Stat(configFile); err != nil {
		log.Fatalf("Error: Cannot open config: %s", err)
	}

	config, err := toml.LoadFile(configFile)
	if err != nil {
		log.Fatalf("Config file %s is malformed: %s", configFile, err)
	}

	addr := getConfigValueDefault(config, "listen", defaultListen)

	url := getConfigValueDefault(config, "letsencrypt.url", defaultURL)
	certsDir := getConfigValueDefault(config, "letsencrypt.certs_dir", defaultCertsDir)
	email := getConfigValue(config, "letsencrypt.email")

	provider = getConfigValue(config, "mail.provider_id")
	imapServer = getConfigValue(config, "mail.imap_server")
	popServer = getConfigValue(config, "mail.pop3_server")
	smtpServer = getConfigValue(config, "mail.smtp_server")

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

func getConfigValueDefault(config *toml.Tree, key string, defaultVal string) string {
	val := config.Get(key)
	if val == nil {
		return defaultVal
	}

	return val.(string)
}

func getConfigValue(config *toml.Tree, key string) string {
	val := config.Get(key)
	if val == nil {
		log.Fatalf("Invalid configuration file: Mandatory setting '%s' is missing", key)
	}

	return val.(string)
}
