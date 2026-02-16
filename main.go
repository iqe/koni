package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	versionFlag    = flag.Bool("V", false, "Print version and exit")
	configFileFlag = flag.String("c", "koni.conf", "Path to configuration file")
	version        = "undefined" // updated during release build

	templateSet *pongo2.TemplateSet
)

const (
	defaultDebug       = false
	defaultListenHTTP  = "127.0.0.1:4080"
	defaultListenHTTPS = "127.0.0.1:4443"
	defaultURL         = "https://acme-staging-v02.api.letsencrypt.org/directory"
	defaultCertsDir    = "."
)

func main() {
	// Remove date + time from logging output (systemd adds those for us)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	flag.Parse()

	if *versionFlag {
		log.Printf("koni - version %s\n", version)
		os.Exit(0)
	}

	config := loadConfigFile(*configFileFlag)

	templateSet = pongo2.NewSet("templates", pongo2.MustNewLocalFileSystemLoader("templates"))

	r := chi.NewRouter()
	r.Use(apacheLogHandler)
	r.Use(middleware.Recoverer)

	if config.debug {
		r.Use(debugLogHandler)
	}

	// Mozilla autoconfig
	r.Get("/mail/config-v1.1.xml", autoconfigHandler(config))
	r.Get("/.well-known/autoconfig/mail/config-v1.1.xml", autoconfigHandler(config))

	// Microsoft autodiscover v1
	autodiscoverXML := autodiscoverxmlHandler(config)
	r.Get("/autodiscover/autodiscover.xml", autodiscoverXML)
	r.Post("/autodiscover/autodiscover.xml", autodiscoverXML)
	r.Get("/Autodiscover/Autodiscover.xml", autodiscoverXML)
	r.Post("/Autodiscover/Autodiscover.xml", autodiscoverXML)

	// Microsoft autodiscover JSON
	r.Get("/autodiscover/autodiscover.json", autodiscoverjsonHandler())

	// Apple iOS mobileconfig
	r.Get("/mobileconfig.xml", mobileconfigHandler(config))

	// Let's Encrypt autocert via tls-alpn-01 challenge
	// See https://tools.ietf.org/html/draft-ietf-acme-tls-alpn-01

	manager := buildAutocertManager(config.url, config.email, config.certsDir)

	// Handler to redirect HTTP to HTTPS
	redirectMux := http.NewServeMux()
	redirectMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	})

	s := &http.Server{
		Addr:         config.listenHTTPS,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second, // Enough time to handle Let's Encrypt challenge on first request for any domain
		IdleTimeout:  120 * time.Second,
		Handler:      r,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1", "acme-tls/1"},
			MinVersion: tls.VersionTLS12, GetCertificate: manager.GetCertificate},
	}

	log.Printf("Starting koni %s...\n", version)
	log.Printf("Let's Encrypt URL: %s\n", config.url)
	log.Printf("Certificate cache directory: %s\n", config.certsDir)

	log.Printf("SMTP server: %s\n", config.smtpServer)
	log.Printf("IMAP server: %s\n", config.imapServer)
	log.Printf("POP3 server: %s\n", config.popServer)

	log.Printf("HTTP server listening on %s\n", config.listenHTTP)
	go func() {
		// This handles ACME http-01 challenges and additionally serves all content over HTTP
		err := http.ListenAndServe(config.listenHTTP, manager.HTTPHandler(redirectMux))

		if err != nil {
			log.Fatalf("Failed to listen on %s: %v\n", config.listenHTTP, err)
		}
	}()

	log.Printf("HTTPS server listening on %s\n", config.listenHTTPS)
	log.Println(s.ListenAndServeTLS("", ""))
}

func renderTemplate(w http.ResponseWriter, name string, status int, ctx pongo2.Context) {
	tpl, err := templateSet.FromFile(name + ".xml.j2")
	if err != nil {
		log.Printf("koni: Failed to load template %s: %v\n", name, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(status)
	if err := tpl.ExecuteWriter(ctx, w); err != nil {
		log.Printf("koni: Failed to render template %s: %v\n", name, err)
	}
}
