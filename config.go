package main

import (
	"bytes"
	"errors"
	"fmt"
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
	Debug       bool            `toml:"debug"`
	ListenHTTP  string          `toml:"listen_http"`
	ListenHTTPS string          `toml:"listen_https"`
	LetsEncrypt tomlLetsEncrypt `toml:"letsencrypt"`
	Mail        tomlMail        `toml:"mail"`
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

func loadConfigFile(configFile string) (koniConfig, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return koniConfig{}, fmt.Errorf("failed to open config file: %w", err)
	}

	var cfg tomlConfig
	decoder := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		var strictErr *toml.StrictMissingError
		if errors.As(err, &strictErr) {
			return koniConfig{}, fmt.Errorf("config file %s contains unknown settings:\n%s", configFile, strictErr.String())
		}
		return koniConfig{}, fmt.Errorf("config file %s is malformed: %w", configFile, err)
	}

	for _, setting := range []struct{ key, val string }{
		{"mail.provider_id", cfg.Mail.ProviderID},
		{"mail.imap_server", cfg.Mail.IMAPServer},
		{"mail.pop3_server", cfg.Mail.POP3Server},
		{"mail.smtp_server", cfg.Mail.SMTPServer},
	} {
		if setting.val == "" {
			return koniConfig{}, fmt.Errorf("invalid configuration file: mandatory setting '%s' is missing", setting.key)
		}
	}

	return koniConfig{
		debug:       cfg.Debug,
		listenHTTP:  stringDefault(cfg.ListenHTTP, defaultListenHTTP),
		listenHTTPS: stringDefault(cfg.ListenHTTPS, defaultListenHTTPS),
		url:         stringDefault(cfg.LetsEncrypt.URL, defaultURL),
		certsDir:    stringDefault(cfg.LetsEncrypt.CertsDir, defaultCertsDir),
		email:       cfg.LetsEncrypt.Email,
		provider:    cfg.Mail.ProviderID,
		imapServer:  cfg.Mail.IMAPServer,
		popServer:   cfg.Mail.POP3Server,
		smtpServer:  cfg.Mail.SMTPServer,
	}, nil
}

func stringDefault(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
