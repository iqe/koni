
VERSION := $(shell git describe --always --long --dirty)

RELEASE_DIR := koni-$(VERSION)
RELEASE_FILE := koni-$(VERSION).tar.gz

.PHONY: release deps
release:
		mkdir -p release
		cd release && rm -rf release/$(RELEASE_DIR) || true
		go build -o release/$(RELEASE_DIR)/koni -v -ldflags="-X main.version=$(VERSION)"
		cp koni.conf koni.service release/$(RELEASE_DIR)/
		cd release && mkdir -p $(RELEASE_DIR)/certs
		cd release && tar czf $(RELEASE_FILE) $(RELEASE_DIR)

deps:
	dep ensure
