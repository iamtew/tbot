# Build settings for tbot.
# This Makefile is designed to work on Windows and Unix-like systems.
BINARY := tbot
PKG := github.com/iamtew/tbot
GOFILES := $(wildcard *.go)
GO := go

ifeq ($(OS),Windows_NT)
	EXE := .exe
	RUNCMD := cmd /c
	RUNBIN := .\\$(BINARY)$(EXE)
	RM := cmd /c del /Q /F $(BINARY)$(EXE)
	# Windows build info via PowerShell
	BUILD_TIME := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss'")
	GIT_COMMIT := $(shell git rev-parse --short HEAD 2>nul || echo "unknown")
	GIT_STATUS := $(shell if git diff-index --quiet HEAD -- 2>nul; then echo "clean"; else echo "dirty"; fi)
	GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>nul || echo "unknown")
	GITHUB_LINK := $(shell git config --get remote.origin.url 2>nul || echo "unknown")
else
	EXE :=
	RUNCMD :=
	RUNBIN := ./$(BINARY)
	RM := rm -f $(BINARY)$(EXE)
	# Unix build info via shell
	BUILD_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
	GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
	GIT_STATUS := $(shell if git diff-index --quiet HEAD -- 2>/dev/null; then echo "clean"; else echo "dirty"; fi)
	GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
	GITHUB_LINK := $(shell git config --get remote.origin.url 2>/dev/null || echo "unknown")
endif

BIN := $(BINARY)$(EXE)
LDFLAGS := -X "main.buildTime=$(BUILD_TIME)" -X "main.gitCommit=$(GIT_COMMIT)" -X "main.gitStatus=$(GIT_STATUS)" -X "main.gitBranch=$(GIT_BRANCH)" -X "main.githubLink=$(GITHUB_LINK)"

.PHONY: all build test fmt fmt-check vet tidy clean install run example

all: build

build: fmt-check $(GOFILES)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BIN) $(PKG)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

fmt-check:
	# Format the code and verify formatting in one step.
	$(GO) fmt ./...

vet:
	go vet ./...

tidy:
	$(GO) mod tidy

clean:
	$(RM)

install:
	$(GO) install $(PKG)

run: build
	$(RUNCMD) $(RUNBIN)

example:
	$(RUNCMD) $(RUNBIN) -e tbot.example.toml

clean:
	-$(RM)
