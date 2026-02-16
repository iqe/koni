package main

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type koniConfig struct {
	debug bool

	listenHTTP  string
	listenHTTPS string

	url      string
	certsDir string
	email    string

	provider   string
	imapServer string
	popServer  string
	smtpServer string
}

type tomlConfig struct {
	Debug       string           `toml:"debug"`
	ListenHTTP  string           `toml:"listen_http"`
	ListenHTTPS string           `toml:"listen_https"`
	LetsEncrypt tomlLetsEncrypt  `toml:"letsencrypt"`
	Mail        tomlMail         `toml:"mail"`
}

type tomlLetsEncrypt struct {
	URL      string `toml:"url"`
	CertsDir string `toml:"certs_dir"`
	Email    string `toml:"email"`
}

type tomlMail struct {
	ProviderID string `toml:"provider_id"`
	IMAPServer string `toml:"imap_server"`
	POP3Server string `toml:"pop3_server"`
	SMTPServer string `toml:"smtp_server"`
}

func loadConfigFile(configFile string) koniConfig {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}

	var cfg tomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Config file %s is malformed: %s", configFile, err)
	}

	requireNonEmpty := func(val, key string) string {
		if val == "" {
			log.Fatalf("Invalid configuration file: Mandatory setting '%s' is missing", key)
		}
		return val
	}

	return koniConfig{
		debug:       stringDefault(cfg.Debug, boolToString(defaultDebug)) == "yes",
		listenHTTP:  stringDefault(cfg.ListenHTTP, defaultListenHTTP),
		listenHTTPS: stringDefault(cfg.ListenHTTPS, defaultListenHTTPS),
		url:         stringDefault(cfg.LetsEncrypt.URL, defaultURL),
		certsDir:    stringDefault(cfg.LetsEncrypt.CertsDir, defaultCertsDir),
		email:       requireNonEmpty(cfg.LetsEncrypt.Email, "letsencrypt.email"),
		provider:    requireNonEmpty(cfg.Mail.ProviderID, "mail.provider_id"),
		imapServer:  requireNonEmpty(cfg.Mail.IMAPServer, "mail.imap_server"),
		popServer:   requireNonEmpty(cfg.Mail.POP3Server, "mail.pop3_server"),
		smtpServer:  requireNonEmpty(cfg.Mail.SMTPServer, "mail.smtp_server"),
	}
}

func stringDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func boolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
