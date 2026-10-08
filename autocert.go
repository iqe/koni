package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	hostRegex = regexp.MustCompile(`(?i)^(autoconfig|autodiscover)\..+$`) // case insensitive
)

// checkCertsDir verifies that the certificate cache directory exists (creating it
// if needed) and is writable. autocert fails every certificate request if it cannot
// store its account key, which would otherwise only show up on the first TLS handshake.
func checkCertsDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".koni-write-test-*")
	if err != nil {
		return err
	}
	return errors.Join(f.Close(), os.Remove(f.Name()))
}

func buildAutocertManager(letsEncryptURL string, email string, certsDir string) *autocert.Manager {
	myHostname, err := os.Hostname()
	if err != nil {
		fatal("Failed to get local hostname from OS", "error", err)
	}

	hostPolicy := func(ctx context.Context, host string) error {
		// Match both autoconfig.*/autodiscover.* and the real hostname because clients
		// handle CNAMEs differently. Given the following CNAME entry:
		//
		// CNAME autoconfig.userdomain.com -> koniserver.provider.com
		//
		// * Thunderbird uses koniserver.provider.com in TLS SNI
		// * Curl uses autoconfig.userdomain.com in TLS SNI
		//
		if hostRegex.MatchString(host) || host == myHostname {
			return nil
		}
		return fmt.Errorf("koni: Hostname %s not allowed by host policy", host)
	}

	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: hostPolicy,
		Cache:      autocert.DirCache(certsDir),
		Email:      email,
		Client:     &acme.Client{DirectoryURL: letsEncryptURL},
	}
}
