BINDIR := bin
VERSION := $(shell git describe --tags --always --dirty)
GOVERSION := $(shell go version)
LDFLAGS := -X 'main.version=$(VERSION)' \
           -X 'main.goversion=$(GOVERSION)'
BUILD_GOOS ?= $(shell go env GOOS)
BUILD_GOARCH ?= $(shell go env GOARCH)

CHECKSUM_FILE := checksums.txt

$(BINDIR):
	install -d $@


.PHONY: build
build: $(BINDIR)
	GOOS=$(BUILD_GOOS) GOARCH=$(BUILD_GOARCH) go build -ldflags "$(LDFLAGS)" -o bin/pd pd.go

.PHONY: build-standalone
build-standalone: build
	mv bin/pd pd-$(VERSION).$(BUILD_GOOS).$(BUILD_GOARCH)
	shasum -a 256 pd-$(VERSION).$(BUILD_GOOS).$(BUILD_GOARCH) >> $(CHECKSUM_FILE)

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" .

.DEFAULT_GOAL := build