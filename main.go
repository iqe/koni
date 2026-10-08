package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/acme/autocert"
)

var (
	versionFlag    = flag.Bool("V", false, "Print version and exit")
	configFileFlag = flag.String("c", "koni.conf", "Path to configuration file")
	version        = "undefined" // updated during release build
)

const (
	contentTypeXML          = "text/xml; charset=utf-8"
	contentTypeMobileconfig = "application/x-apple-aspen-config"

	defaultListenHTTP  = "127.0.0.1:4080"
	defaultListenHTTPS = "127.0.0.1:4443"
	defaultURL         = stagingURL
	defaultCertsDir    = "."

	stagingURL = "https://acme-staging-v02.api.letsencrypt.org/directory"

	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second // Enough time to handle Let's Encrypt challenge on first request for any domain
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// Remove date + time from logging output (systemd adds those for us)
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	flag.Parse()

	if *versionFlag {
		log.Printf("koni - version %s\n", version)
		os.Exit(0)
	}

	config, err := loadConfigFile(*configFileFlag)
	if err != nil {
		log.Fatalf("koni: %v\n", err)
	}

	// Let's Encrypt autocert via tls-alpn-01 and http-01 challenges
	manager := buildAutocertManager(config.url, config.email, config.certsDir)

	httpServer, httpsServer := buildServers(config, buildRouter(config), manager)

	log.Printf("Starting koni %s...\n", version)
	log.Printf("Let's Encrypt URL: %s\n", config.url)
	if config.url == stagingURL {
		log.Printf("WARNING: Using the Let's Encrypt STAGING environment. Clients will not trust the certificates. Set letsencrypt.url for production use.\n")
	}
	log.Printf("Certificate cache directory: %s\n", config.certsDir)

	log.Printf("SMTP server: %s\n", config.smtpServer)
	log.Printf("IMAP server: %s\n", config.imapServer)
	log.Printf("POP3 server: %s\n", config.popServer)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := serve(ctx, httpServer, httpsServer); err != nil {
		log.Fatalf("koni: %v\n", err)
	}
	log.Println("koni stopped")
}

func buildRouter(config koniConfig) http.Handler {
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

	return r
}

func buildServers(config koniConfig, handler http.Handler, manager *autocert.Manager) (*http.Server, *http.Server) {
	// Handler to redirect HTTP to HTTPS
	redirectMux := http.NewServeMux()
	redirectMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	})

	// This handles ACME http-01 challenges and redirects everything else to HTTPS
	httpServer := &http.Server{
		Addr:              config.listenHTTP,
		Handler:           manager.HTTPHandler(redirectMux),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// TLSConfig sets GetCertificate and NextProtos (including acme-tls/1 for tls-alpn-01)
	tlsConfig := manager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12

	httpsServer := &http.Server{
		Addr:              config.listenHTTPS,
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	return httpServer, httpsServer
}

// serve runs both servers until ctx is cancelled or one of them fails,
// then shuts down both gracefully.
func serve(ctx context.Context, httpServer, httpsServer *http.Server) error {
	errc := make(chan error, 2)

	go func() {
		log.Printf("HTTP server listening on %s\n", httpServer.Addr)
		errc <- httpServer.ListenAndServe()
	}()
	go func() {
		log.Printf("HTTPS server listening on %s\n", httpsServer.Addr)
		errc <- httpsServer.ListenAndServeTLS("", "")
	}()

	var serveErr error
	select {
	case <-ctx.Done():
		log.Println("Shutting down...")
	case serveErr = <-errc:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := errors.Join(httpServer.Shutdown(shutdownCtx), httpsServer.Shutdown(shutdownCtx))
	return errors.Join(serveErr, shutdownErr)
}
