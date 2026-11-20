# qdis - An efficient in-memory cache in Go
# Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved.
# Use of this source code is governed by GPL-3. See LICENSE.

GO         ?= go
GOLANGCI   ?= golangci-lint
PKG        ?= ./...
BINDIR     := build
BINARY     := $(BINDIR)/qdis
CMD        := ./cmd/qdis
VERSION    := 0.1.0

GOFILES    := $(shell find . -name '*.go' -not -path './vendor/*' -not -path './.git/*' 2>/dev/null)

.DEFAULT_GOAL := help

.PHONY: all build test race fmt fmtcheck vet check ci clean line version help

all: build

build:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@mkdir -p $(BINDIR)
	@$(GO) build -trimpath -o $(BINARY) $(CMD)
	@echo "  GO    $(BINARY)"
endif

test:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@$(GO) test $(PKG)
endif

race:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@$(GO) test -race $(PKG)
endif

fmt:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@$(GO) fmt $(PKG)
	@gofmt -s -w $(GOFILES)
	@echo "  FMT   ok"
endif

fmtcheck:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@unformatted=$$(gofmt -s -l $(GOFILES)); \
	if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi; \
	echo "  FMT   check ok"
endif

vet:
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@$(GO) vet $(PKG)
	@echo "  VET   ok"
endif

check: fmt vet
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@command -v $(GOLANGCI) >/dev/null 2>&1 || { \
		echo "  ERR   $(GOLANGCI) not found. Install:"; \
		echo "        go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8"; \
		exit 1; \
	}
	@$(GOLANGCI) run $(PKG)
	@echo "  LINT  ok"
endif

ci: fmtcheck vet
ifeq ($(strip $(GOFILES)),)
	@echo "  skip  no Go files yet"
else
	@command -v $(GOLANGCI) >/dev/null 2>&1 || { \
		echo "  ERR   $(GOLANGCI) not found"; \
		exit 1; \
	}
	@$(GOLANGCI) run $(PKG)
	@$(GO) test $(PKG)
	@echo "  CI    ok"
endif

line:
	@files=$$(find . -name '*.go' ! -name '*_test.go' ! -path './vendor/*' ! -path './.git/*' 2>/dev/null | sort); \
	if [ -z "$$files" ]; then echo "  skip  no Go files yet"; else awk -f scripts/line.awk $$files; fi

clean:
	@rm -rf $(BINDIR)
	@echo "  CLEAN $(BINDIR)"

version:
	@echo "qdis v$(VERSION)"

help:
	@echo "qdis - An efficient in-memory cache in Go v$(VERSION)"
	@echo "Copyright (C) 2026 - 2027 Yixiang Qiu. All rights reserved."
	@echo ""
	@echo "USAGE:"
	@echo "\tmake all                 - build the server binary"
	@echo "\tmake build               - build the server binary"
	@echo "\tmake test                - run the test suite"
	@echo "\tmake race                - run tests with the race detector"
	@echo "\tmake fmt                 - format Go sources"
	@echo "\tmake fmtcheck            - fail if sources are unformatted"
	@echo "\tmake vet                 - run go vet"
	@echo "\tmake check               - format, vet, and lint"
	@echo "\tmake ci                  - fmtcheck, vet, lint, and test"
	@echo "\tmake clean               - remove the build directory"
	@echo "\tmake line                - count implementation lines"
	@echo "\tmake version             - print the project version"
	@echo "\tmake help                - show this help message"
	@echo ""
