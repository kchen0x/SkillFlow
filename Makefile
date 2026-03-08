# Detect OS and set appropriate path
ifeq ($(OS),Windows_NT)
    SHELL := cmd.exe
    .SHELLFLAGS := /C
    WAILS := $(shell where wails.exe 2>nul || echo $(USERPROFILE)\go\bin\wails.exe)
else
    WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)
endif

APP_DIR := cmd/skillflow
APP_DIR_WIN := $(subst /,\,$(APP_DIR))

.PHONY: all dev build test tidy generate install-frontend clean help

all: build

## dev: Run in dev mode with hot-reload (Go + frontend)
dev:
	cd $(APP_DIR) && $(WAILS) dev

## build: Build production binary
build:
	cd $(APP_DIR) && $(WAILS) build
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
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
	cd $(APP_DIR) && $(WAILS) generate module

## install-frontend: Install frontend npm dependencies
install-frontend:
	cd $(APP_DIR)/frontend && npm install

## clean: Remove build artifacts
clean:
ifeq ($(OS),Windows_NT)
	if exist "$(APP_DIR_WIN)\build\bin" rmdir /s /q "$(APP_DIR_WIN)\build\bin"
	if exist "$(APP_DIR_WIN)\frontend\dist" rmdir /s /q "$(APP_DIR_WIN)\frontend\dist"
	if exist "$(APP_DIR_WIN)\frontend\package.json.md5" del /f /q "$(APP_DIR_WIN)\frontend\package.json.md5"
else
	rm -rf $(APP_DIR)/build/bin
	rm -rf $(APP_DIR)/frontend/dist
	rm -f $(APP_DIR)/frontend/package.json.md5
endif

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
