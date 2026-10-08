package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestHostPolicy(t *testing.T) {
	manager := buildAutocertManager(
		"https://acme-staging-v02.api.letsencrypt.org/directory",
		"test@example.com",
		t.TempDir(),
	)

	ctx := context.Background()

	tests := []struct {
		host    string
		allowed bool
	}{
		{"autoconfig.example.com", true},
		{"autodiscover.example.com", true},
		{"AUTOCONFIG.example.com", true},
		{"AUTODISCOVER.example.com", true},
		{"autoconfig.sub.example.com", true},
		{"mail.example.com", false},
		{"example.com", false},
		{"notautoconfig.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			err := manager.HostPolicy(ctx, tt.host)
			if tt.allowed && err != nil {
				t.Errorf("HostPolicy(%q) returned error: %v, want nil", tt.host, err)
			}
			if !tt.allowed && err == nil {
				t.Errorf("HostPolicy(%q) returned nil, want error", tt.host)
			}
		})
	}
}

func TestCheckCertsDir(t *testing.T) {
	t.Run("writable", func(t *testing.T) {
		dir := t.TempDir()
		if err := checkCertsDir(dir); err != nil {
			t.Fatalf("checkCertsDir: %v", err)
		}
		entries, _ := os.ReadDir(dir)
		if len(entries) != 0 {
			t.Errorf("test file was not removed: %v", entries)
		}
	})

	t.Run("created if missing", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "certs")
		if err := checkCertsDir(dir); err != nil {
			t.Fatalf("checkCertsDir: %v", err)
		}
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("directory was not created: %v", err)
		}
	})

	t.Run("not writable", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can write to any directory")
		}
		dir := t.TempDir()
		if err := os.Chmod(dir, 0500); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(dir, 0700)
		if err := checkCertsDir(dir); err == nil {
			t.Error("checkCertsDir returned nil, want error")
		}
	})
}
