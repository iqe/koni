package main

import (
	"log"

	toml "github.com/pelletier/go-toml"
)

type koniConfig struct {
	listen string

	url      string
	certsDir string
	email    string

	provider   string
	imapServer string
	popServer  string
	smtpServer string
}

func loadConfigFile(configFile string) koniConfig {
	tomlConfig, err := toml.LoadFile(configFile)
	if err != nil {
		log.Fatalf("Config file %s is malformed: %s", configFile, err)
	}

	return koniConfig{
		listen:     getConfigValueDefault(tomlConfig, "listen", defaultListen),
		url:        getConfigValueDefault(tomlConfig, "letsencrypt.url", defaultURL),
		certsDir:   getConfigValueDefault(tomlConfig, "letsencrypt.certs_dir", defaultCertsDir),
		email:      getConfigValue(tomlConfig, "letsencrypt.email"),
		provider:   getConfigValue(tomlConfig, "mail.provider_id"),
		imapServer: getConfigValue(tomlConfig, "mail.imap_server"),
		popServer:  getConfigValue(tomlConfig, "mail.pop3_server"),
		smtpServer: getConfigValue(tomlConfig, "mail.smtp_server"),
	}
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
