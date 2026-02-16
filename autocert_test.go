package main

import (
	"context"
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
