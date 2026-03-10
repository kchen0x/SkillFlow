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
NODE ?= node
EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,

WAILS_BUILD_TAGS ?=
WAILS_BUILD_LDFLAGS ?= -s -w
BUILD_PLAN_CMD = $(NODE) $(APP_DIR)/tools/build-plan.mjs
NORMALIZE_PROVIDERS = $(strip $(subst $(COMMA),$(SPACE),$(1)))
PROVIDER_BUILD_TAGS = $(strip provider_select $(foreach provider,$(call NORMALIZE_PROVIDERS,$(1)),backup_$(provider)))
WAILS_SKIP_FLAGS = $(shell $(BUILD_PLAN_CMD) plan)
WAILS_BUILD_FLAGS = -trimpath -m -nosyncgomod $(WAILS_SKIP_FLAGS) $(if $(strip $(WAILS_BUILD_TAGS)),-tags "$(WAILS_BUILD_TAGS)") $(if $(strip $(WAILS_BUILD_LDFLAGS)),-ldflags "$(WAILS_BUILD_LDFLAGS)")

.PHONY: all dev build build-cloud test test-cloud tidy generate install-frontend clean help

all: build

## dev: Run in dev mode with hot-reload (Go + frontend)
dev:
	cd $(APP_DIR) && $(WAILS) dev

## build: Build production binary
build:
	cd $(APP_DIR) && $(WAILS) build $(WAILS_BUILD_FLAGS)
	$(BUILD_PLAN_CMD) mark
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## build-cloud: Build with only selected cloud providers, e.g. make build-cloud PROVIDERS="aws,google"
build-cloud:
	$(if $(strip $(PROVIDERS)),,$(error Usage: make build-cloud PROVIDERS="aws,google"))
	cd $(APP_DIR) && $(WAILS) build -trimpath -m -nosyncgomod $(WAILS_SKIP_FLAGS) -tags "$(strip $(WAILS_BUILD_TAGS) $(call PROVIDER_BUILD_TAGS,$(PROVIDERS)))" $(if $(strip $(WAILS_BUILD_LDFLAGS)),-ldflags "$(WAILS_BUILD_LDFLAGS)")
	$(BUILD_PLAN_CMD) mark
ifneq ($(OS),Windows_NT)
	@if [ -f $(APP_DIR)/build/darwin/iconfile.icns ] && [ -d $(APP_DIR)/build/bin/SkillFlow.app ]; then \
		cp $(APP_DIR)/build/darwin/iconfile.icns $(APP_DIR)/build/bin/SkillFlow.app/Contents/Resources/iconfile.icns; \
	fi
endif

## test: Run all Go tests
test:
	go test ./core/...

## test-cloud: Run Go tests with only selected cloud providers, e.g. make test-cloud PROVIDERS="aws,google"
test-cloud:
	$(if $(strip $(PROVIDERS)),,$(error Usage: make test-cloud PROVIDERS="aws,google"))
	go test -tags "$(strip $(WAILS_BUILD_TAGS) $(call PROVIDER_BUILD_TAGS,$(PROVIDERS)))" ./core/...

## tidy: Sync Go module dependencies
tidy:
	go mod tidy

## generate: Regenerate TypeScript bindings after App method changes
generate:
	cd $(APP_DIR) && $(WAILS) generate module
	$(BUILD_PLAN_CMD) mark-bindings

## install-frontend: Install frontend npm dependencies
install-frontend:
	cd $(APP_DIR)/frontend && npm install

## clean: Remove build artifacts
clean:
ifeq ($(OS),Windows_NT)
	if exist "$(APP_DIR_WIN)\build\bin" rmdir /s /q "$(APP_DIR_WIN)\build\bin"
	if exist "$(APP_DIR_WIN)\frontend\dist" rmdir /s /q "$(APP_DIR_WIN)\frontend\dist"
	if exist "$(APP_DIR_WIN)\frontend\.cache" rmdir /s /q "$(APP_DIR_WIN)\frontend\.cache"
	if exist "$(APP_DIR_WIN)\frontend\package.json.md5" del /f /q "$(APP_DIR_WIN)\frontend\package.json.md5"
	if exist "$(APP_DIR_WIN)\.build-cache" rmdir /s /q "$(APP_DIR_WIN)\.build-cache"
else
	rm -rf $(APP_DIR)/build/bin
	rm -rf $(APP_DIR)/frontend/dist
	rm -rf $(APP_DIR)/frontend/.cache
	rm -f $(APP_DIR)/frontend/package.json.md5
	rm -rf $(APP_DIR)/.build-cache
endif

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
