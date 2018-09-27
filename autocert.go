package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

var (
	hostRegex = regexp.MustCompile("(?i)^(autoconfig|autodiscover)\\..+$") // case insensitive
)

func buildAutocertManager(letsEncryptURL string, email string, certsDir string) *autocert.Manager {
	myHostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("koni: failed to get local hostname from OS: %s\n", err)
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
