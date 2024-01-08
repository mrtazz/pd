BINDIR := bin
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.goversion=$(GOVERSION)'
BUILD_GOOS ?= $(shell go env GOOS)
BUILD_GOARCH ?= $(shell go env GOARCH)

RELEASE_ARTIFACTS_DIR := .release_artifacts
CHECKSUM_FILE := $(RELEASE_ARTIFACTS_DIR)/checksums.txt

$(RELEASE_ARTIFACTS_DIR):
	install -d $@

$(BINDIR):
	install -d $@

###
# Man page build tasks
###

BUILD_DAY := $(shell date -u +"%Y-%m-%d")
MANPAGE := docs/man/pd.1
PREFIX ?= "/usr/local"

.PHONY: man
man: $(MANPAGE)

$(MANPAGE): $(MANPAGE).md
	sed "s/VERSION_PLACEHOLDER/${VERSION}/g" $< | \
	 	sed "s/DATE_PLACEHOLDER/${BUILD_DAY}/g" | \
	 	pandoc --standalone -f markdown -t man -o $@

.PHONY: local-install
local-install:
	$(MAKE) install PREFIX=usr/local


.PHONY: build
build: $(BINDIR)
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -ldflags "$(LDFLAGS)" -o bin/pd pd.go

.PHONY: build-standalone
build-standalone: build man $(RELEASE_ARTIFACTS_DIR)
	mv bin/pd $(RELEASE_ARTIFACTS_DIR)/pd-$(VERSION).$(BUILD_GOOS).$(BUILD_GOARCH)
	mv $(MANPAGE) $(RELEASE_ARTIFACTS_DIR)/
	shasum -a 256 $(RELEASE_ARTIFACTS_DIR)/pd-$(VERSION).$(BUILD_GOOS).$(BUILD_GOARCH) >> $(CHECKSUM_FILE)

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" .

.DEFAULT_GOAL := build

.PHONY: github-release
github-release:
	gh release create $(VERSION) --title 'Release $(VERSION)' \
	 	--notes-file docs/releases/$(VERSION).md $(RELEASE_ARTIFACTS_DIR)/*


