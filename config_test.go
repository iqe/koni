package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringDefault(t *testing.T) {
	t.Run("non-empty value", func(t *testing.T) {
		got := stringDefault("value1", "default")
		if got != "value1" {
			t.Errorf("got %q, want %q", got, "value1")
		}
	})

	t.Run("empty returns default", func(t *testing.T) {
		got := stringDefault("", "fallback")
		if got != "fallback" {
			t.Errorf("got %q, want %q", got, "fallback")
		}
	})
}

func TestBoolToString(t *testing.T) {
	if got := boolToString(true); got != "yes" {
		t.Errorf("boolToString(true) = %q, want %q", got, "yes")
	}
	if got := boolToString(false); got != "no" {
		t.Errorf("boolToString(false) = %q, want %q", got, "no")
	}
}

func TestLoadConfigFile(t *testing.T) {
	content := `
debug = "yes"
listen_http = "0.0.0.0:80"
listen_https = "0.0.0.0:443"

[letsencrypt]
url = "https://acme-v02.api.letsencrypt.org/directory"
email = "admin@example.com"
certs_dir = "/var/certs"

[mail]
provider_id = "example.com"
smtp_server = "smtp.example.com"
imap_server = "imap.example.com"
pop3_server = "pop.example.com"
`
	path := filepath.Join(t.TempDir(), "test.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadConfigFile(path)

	if !cfg.debug {
		t.Error("debug: got false, want true")
	}
	if cfg.listenHTTP != "0.0.0.0:80" {
		t.Errorf("listenHTTP = %q, want %q", cfg.listenHTTP, "0.0.0.0:80")
	}
	if cfg.listenHTTPS != "0.0.0.0:443" {
		t.Errorf("listenHTTPS = %q, want %q", cfg.listenHTTPS, "0.0.0.0:443")
	}
	if cfg.url != "https://acme-v02.api.letsencrypt.org/directory" {
		t.Errorf("url = %q, want acme-v02 URL", cfg.url)
	}
	if cfg.email != "admin@example.com" {
		t.Errorf("email = %q, want %q", cfg.email, "admin@example.com")
	}
	if cfg.certsDir != "/var/certs" {
		t.Errorf("certsDir = %q, want %q", cfg.certsDir, "/var/certs")
	}
	if cfg.provider != "example.com" {
		t.Errorf("provider = %q, want %q", cfg.provider, "example.com")
	}
	if cfg.smtpServer != "smtp.example.com" {
		t.Errorf("smtpServer = %q, want %q", cfg.smtpServer, "smtp.example.com")
	}
	if cfg.imapServer != "imap.example.com" {
		t.Errorf("imapServer = %q, want %q", cfg.imapServer, "imap.example.com")
	}
	if cfg.popServer != "pop.example.com" {
		t.Errorf("popServer = %q, want %q", cfg.popServer, "pop.example.com")
	}
}

func TestLoadConfigFileDefaults(t *testing.T) {
	content := `
[letsencrypt]
email = "admin@example.com"

[mail]
provider_id = "example.com"
smtp_server = "smtp.example.com"
imap_server = "imap.example.com"
pop3_server = "pop.example.com"
`
	path := filepath.Join(t.TempDir(), "test.conf")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg := loadConfigFile(path)

	if cfg.debug {
		t.Error("debug: got true, want false (default)")
	}
	if cfg.listenHTTP != defaultListenHTTP {
		t.Errorf("listenHTTP = %q, want default %q", cfg.listenHTTP, defaultListenHTTP)
	}
	if cfg.listenHTTPS != defaultListenHTTPS {
		t.Errorf("listenHTTPS = %q, want default %q", cfg.listenHTTPS, defaultListenHTTPS)
	}
	if cfg.url != defaultURL {
		t.Errorf("url = %q, want default %q", cfg.url, defaultURL)
	}
	if cfg.certsDir != defaultCertsDir {
		t.Errorf("certsDir = %q, want default %q", cfg.certsDir, defaultCertsDir)
	}
}
