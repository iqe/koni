package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
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
	flag.Parse()

	if *versionFlag {
		fmt.Printf("koni - version %s\n", version)
		os.Exit(0)
	}

	logLevel := new(slog.LevelVar)
	slog.SetDefault(slog.New(newLogHandler(os.Stderr, logLevel)))

	config, err := loadConfigFile(*configFileFlag)
	if err != nil {
		fatal("Failed to load configuration", "error", err)
	}
	if config.debug {
		logLevel.Set(slog.LevelDebug)
	}

	// Let's Encrypt autocert via tls-alpn-01 and http-01 challenges
	manager := buildAutocertManager(config.url, config.email, config.certsDir)

	httpServer, httpsServer := buildServers(config, buildRouter(config), manager)

	slog.Info("Starting koni",
		"version", version,
		"letsencrypt_url", config.url,
		"certs_dir", config.certsDir,
		"smtp_server", config.smtpServer,
		"imap_server", config.imapServer,
		"pop3_server", config.popServer,
	)
	if config.url == stagingURL {
		slog.Warn("Using the Let's Encrypt STAGING environment. Clients will not trust the certificates. Set letsencrypt.url for production use.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := serve(ctx, httpServer, httpsServer); err != nil {
		fatal("Server failed", "error", err)
	}
	slog.Info("koni stopped")
}

// newLogHandler creates a text log handler without timestamps (systemd adds those for us)
func newLogHandler(w io.Writer, level slog.Leveler) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	})
}

func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func buildRouter(config koniConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(accessLogHandler)
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
		Handler:           accessLogHandler(manager.HTTPHandler(redirectMux)),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
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
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
	}

	return httpServer, httpsServer
}

// serve runs both servers until ctx is cancelled or one of them fails,
// then shuts down both gracefully.
func serve(ctx context.Context, httpServer, httpsServer *http.Server) error {
	errc := make(chan error, 2)

	go func() {
		slog.Info("HTTP server listening", "addr", httpServer.Addr)
		errc <- httpServer.ListenAndServe()
	}()
	go func() {
		slog.Info("HTTPS server listening", "addr", httpsServer.Addr)
		errc <- httpsServer.ListenAndServeTLS("", "")
	}()

	var serveErr error
	select {
	case <-ctx.Done():
		slog.Info("Shutting down")
	case serveErr = <-errc:
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	shutdownErr := errors.Join(httpServer.Shutdown(shutdownCtx), httpsServer.Shutdown(shutdownCtx))
	return errors.Join(serveErr, shutdownErr)
}
