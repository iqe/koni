package main

import (
	"context"
	"fmt"
	"regexp"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	hostRegex = regexp.MustCompile("(?i)^(autoconfig|autodiscover)\\..+$") // case insensitive
)

func buildAutocertManager(letsEncryptURL string, email string, certsDir string) *autocert.Manager {
	hostPolicy := func(ctx context.Context, host string) error {
		if hostRegex.MatchString(host) {
			return nil
		}
		return fmt.Errorf("acme/autocert: invalid hostname: %s, only autoconfig/autodiscover hosts are allowed", host)
	}

	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(certsDir),
		Email:      email,
		Client:     &acme.Client{DirectoryURL: letsEncryptURL},
	}
}
