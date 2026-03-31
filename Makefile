.PHONY: build build-linux build-windows build-all clean test

BINARY_NAME := port-scan
MAIN_PATH := ./cmd/port-scan
VERSION := $(shell git describe --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Go parameters
GOCMD ?= go
GOARCH_AMD64 := amd64
GOOS_LINUX := linux
GOOS_WINDOWS := windows

# Output directories
DIST_DIR := dist

# Default target
build: build-all

## build: Build all targets (linux and windows x64)
build-all: build-linux build-windows

## build-linux: Build Linux x64 binary
build-linux:
	@echo "Building Linux x64..."
	@mkdir -p $(DIST_DIR)/linux
	$(GOCMD) build $(LDFLAGS) -o $(DIST_DIR)/linux/$(BINARY_NAME) $(MAIN_PATH)

## build-windows: Build Windows x64 binary
build-windows:
	@echo "Building Windows x64..."
	@mkdir -p $(DIST_DIR)/windows
	GOOS=$(GOOS_WINDOWS) GOARCH=$(GOARCH_AMD64) $(GOCMD) build $(LDFLAGS) -o $(DIST_DIR)/windows/$(BINARY_NAME).exe $(MAIN_PATH)

## clean: Remove build artifacts
clean:
	rm -rf $(DIST_DIR)

## test: Run tests
test:
	$(GOCMD) test -race -shuffle=on ./...

## lint: Run linter
lint:
	$(GOCMD) vet ./...
