package main

import (
	"context"
	"net"
	"net/http"
	"slices"
	"testing"
	"time"
)

func TestBuildServers(t *testing.T) {
	config := koniConfig{listenHTTP: "127.0.0.1:0", listenHTTPS: "127.0.0.1:0"}
	manager := buildAutocertManager(stagingURL, "", t.TempDir())

	httpServer, httpsServer := buildServers(config, http.NotFoundHandler(), manager)

	for _, s := range []*http.Server{httpServer, httpsServer} {
		if s.ReadHeaderTimeout == 0 || s.ReadTimeout == 0 || s.WriteTimeout == 0 || s.IdleTimeout == 0 {
			t.Errorf("server %s has unset timeouts", s.Addr)
		}
	}
	if !slices.Contains(httpsServer.TLSConfig.NextProtos, "acme-tls/1") {
		t.Errorf("NextProtos = %v, missing acme-tls/1", httpsServer.TLSConfig.NextProtos)
	}
	if httpsServer.TLSConfig.GetCertificate == nil {
		t.Error("GetCertificate not set")
	}
}

func TestServeShutsDownOnCancel(t *testing.T) {
	config := koniConfig{listenHTTP: "127.0.0.1:0", listenHTTPS: "127.0.0.1:0"}
	manager := buildAutocertManager(stagingURL, "", t.TempDir())
	httpServer, httpsServer := buildServers(config, http.NotFoundHandler(), manager)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serve(ctx, httpServer, httpsServer) }()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("serve returned %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after cancel")
	}
}

func TestServeFailsWhenPortInUse(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	config := koniConfig{listenHTTP: l.Addr().String(), listenHTTPS: "127.0.0.1:0"}
	manager := buildAutocertManager(stagingURL, "", t.TempDir())
	httpServer, httpsServer := buildServers(config, http.NotFoundHandler(), manager)

	done := make(chan error, 1)
	go func() { done <- serve(context.Background(), httpServer, httpsServer) }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("serve returned nil, want error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return on listen error")
	}
}
