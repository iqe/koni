package main

import (
	"log"
	"os"

	toml "github.com/pelletier/go-toml"
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

func loadConfigFile(configFile string) koniConfig {
	if _, err := os.Stat(configFile); err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}

	tomlConfig, err := toml.LoadFile(configFile)
	if err != nil {
		log.Fatalf("Config file %s is malformed: %s", configFile, err)
	}

	return koniConfig{
		debug:       getBoolConfigValueDefault(tomlConfig, "debug", defaultDebug),
		listenHTTP:  getConfigValueDefault(tomlConfig, "listen_http", defaultListenHTTP),
		listenHTTPS: getConfigValueDefault(tomlConfig, "listen_https", defaultListenHTTPS),
		url:         getConfigValueDefault(tomlConfig, "letsencrypt.url", defaultURL),
		certsDir:    getConfigValueDefault(tomlConfig, "letsencrypt.certs_dir", defaultCertsDir),
		email:       getConfigValue(tomlConfig, "letsencrypt.email"),
		provider:    getConfigValue(tomlConfig, "mail.provider_id"),
		imapServer:  getConfigValue(tomlConfig, "mail.imap_server"),
		popServer:   getConfigValue(tomlConfig, "mail.pop3_server"),
		smtpServer:  getConfigValue(tomlConfig, "mail.smtp_server"),
	}
}

func getConfigValueDefault(config *toml.Tree, key string, defaultVal string) string {
	val := config.Get(key)
	if val == nil {
		return defaultVal
	}

	s, ok := val.(string)
	if !ok {
		log.Fatalf("Invalid configuration file: Setting '%s' must be a string", key)
	}

	return s
}

func getConfigValue(config *toml.Tree, key string) string {
	val := config.Get(key)
	if val == nil {
		log.Fatalf("Invalid configuration file: Mandatory setting '%s' is missing", key)
	}

	s, ok := val.(string)
	if !ok {
		log.Fatalf("Invalid configuration file: Setting '%s' must be a string", key)
	}

	return s
}

func getBoolConfigValueDefault(config *toml.Tree, key string, defaultVal bool) bool {
	defaultStringValue := "no"
	if defaultVal {
		defaultStringValue = "yes"
	}

	return getConfigValueDefault(config, key, defaultStringValue) == "yes"
}
