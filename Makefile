GOMODNAME := $(shell grep 'module' go.mod | sed -e 's/^module //')
SOURCES := $(shell find . -name "*.go" -or -name "go.mod" -or -name "go.sum" \
	-or -name "Makefile")

# Verbose output
ifdef VERBOSE
V = -v
endif

#
# Environment
#

BINDIR := bin
TOOLDIR := $(BINDIR)/tools

# Global environment variables for all targets
SHELL ?= /bin/bash
SHELL := env \
	GO111MODULE=on \
	GOBIN=$(CURDIR)/$(TOOLDIR) \
	CGO_ENABLED=0 \
	PATH='$(CURDIR)/$(BINDIR):$(CURDIR)/$(TOOLDIR):$(PATH)' \
	$(SHELL)

#
# Defaults
#

# Default target
.DEFAULT_GOAL := build

#
# Tools
#

# external tool
define tool # 1: binary-name, 2: go-import-path
TOOLS += $(TOOLDIR)/$(1)

$(TOOLDIR)/$(1): Makefile
	GOBIN="$(CURDIR)/$(TOOLDIR)" go install "$(2)"
endef

$(eval $(call tool,godoc,golang.org/x/tools/cmd/godoc@latest))
$(eval $(call tool,gofumpt,mvdan.cc/gofumpt@latest))
$(eval $(call tool,golangci-lint,github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59))
$(eval $(call tool,gomod,github.com/Helcaraxan/gomod@latest))
$(eval $(call tool,mockgen,github.com/golang/mock/mockgen@v1.6.0))

.PHONY: tools
tools: $(TOOLS)

#
# Build
#

NAME = sshush
BINARY = dist/${NAME}
LDFLAGS := -w -s

VERSION ?= $(shell git describe --tags --exact 2>/dev/null)
GIT_SHA ?= $(shell git rev-parse HEAD 2>/dev/null)

DEV_VERSION = 0.0.0-dev
ifeq ($(VERSION),)
	VERSION = $(DEV_VERSION)
endif

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build $(V) -o "$@" -ldflags "$(LDFLAGS) \
		-X main.version=$(VERSION) \
		-X main.commit=$(GIT_SHA)"

.PHONY: build
build: $(BINARY)

.PHONY: run
run: $(BINARY)
	$(BINARY)

#
# Development
#

BENCH ?= .
TESTARGS ?=

.PHONY: clean
clean:
	rm -f $(TOOLS)
	rm -f ./coverage.out ./go.mod.tidy-check ./go.sum.tidy-check

.PHONY: test
test:
	CGO_ENABLED=0 go test $(V) -count=1 $(TESTARGS) ./...

.PHONY: test-deps
test-deps:
	go test all

.PHONY: lint
lint: $(TOOLDIR)/golangci-lint
	golangci-lint $(V) run

.PHONY: format
format: $(TOOLDIR)/goimports $(TOOLDIR)/gofumpt
	goimports -w . && gofumpt -w .

.SILENT: bench
.PHONY: bench
bench:
	go test $(V) -count=1 -bench=$(BENCH) $(TESTARGS) ./...

#
# Code Generation
#

.PHONY: generate
generate: $(TOOLDIR)/mockgen
	go generate ./...

.PHONY: check-generate
check-generate: $(TOOLDIR)/mockgen
	$(eval CHKDIR := $(shell mktemp -d))
	cp -av . "$(CHKDIR)"
	make -C "$(CHKDIR)/" generate
	( diff -rN . "$(CHKDIR)" && rm -rf "$(CHKDIR)" ) || \
	( rm -rf "$(CHKDIR)" && exit 1 )

#
# Coverage
#

.PHONY: cov
cov: coverage.out

.PHONY: cov-html
cov-html: coverage.out
	go tool cover -html=./coverage.out

.PHONY: cov-func
cov-func: coverage.out
	go tool cover -func=./coverage.out

coverage.out: $(SOURCES)
	go test $(V) -covermode=count -coverprofile=./coverage.out ./...

#
# Dependencies
#

.PHONY: deps
deps:
	go mod download

.PHONY: deps-update
deps-update:
	go get -u -t ./...

.PHONY: deps-analyze
deps-analyze: $(TOOLDIR)/gomod
	gomod analyze

.PHONY: tidy
tidy:
	go mod tidy $(V)

.PHONY: verify
verify:
	go mod verify

.SILENT: check-tidy
.PHONY: check-tidy
check-tidy:
	cp go.mod go.mod.tidy-check
	cp go.sum go.sum.tidy-check
	go mod tidy
	( \
		diff go.mod go.mod.tidy-check && \
		diff go.sum go.sum.tidy-check && \
		rm -f go.mod go.sum && \
		mv go.mod.tidy-check go.mod && \
		mv go.sum.tidy-check go.sum \
	) || ( \
		rm -f go.mod go.sum && \
		mv go.mod.tidy-check go.mod && \
		mv go.sum.tidy-check go.sum; \
		exit 1 \
	)

#
# Documentation
#

# Serve docs
.PHONY: docs
docs: $(TOOLDIR)/godoc
	$(info serviing docs on http://127.0.0.1:6060/pkg/$(GOMODNAME)/)
	@godoc -http=127.0.0.1:6060
