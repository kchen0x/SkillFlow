# Detect OS and set appropriate path
ifeq ($(OS),Windows_NT)
    WAILS := $(shell where wails 2>/dev/null || echo $(USERPROFILE)/go/bin/wails.exe)
else
    WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)
endif

.PHONY: all dev build test tidy generate install-frontend clean help

all: build

## dev: Run in dev mode with hot-reload (Go + frontend)
dev:
	$(WAILS) dev
## build: Build production binary
build:
	$(WAILS) build
ifneq ($(OS),Windows_NT)
	@if [ -f build/darwin/iconfile.icns ] && [ -d build/bin/SkillFlow.app ]; then \
		cp build/darwin/iconfile.icns build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## test: Run all Go tests
test:
	go test ./core/...

## tidy: Sync Go module dependencies
tidy:
	go mod tidy

## generate: Regenerate TypeScript bindings after App method changes
generate:
	$(WAILS) generate module

## install-frontend: Install frontend npm dependencies
install-frontend:
	cd frontend && npm install

## clean: Remove build artifacts
clean:
ifeq ($(OS),Windows_NT)
	if exist build\bin rmdir /s /q build\bin
else
	rm -rf build/bin
endif

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
