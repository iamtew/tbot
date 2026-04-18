BINARY := tbot
PKG := github.com/iamtew/tbot
GOFILES := $(wildcard *.go)
GO := go

ifeq ($(OS),Windows_NT)
	EXE := .exe
	RUNCMD := cmd /c
	RUNBIN := .\\$(BINARY)$(EXE)
	RM := cmd /c del /Q $(BINARY)$(EXE)
else
	EXE :=
	RUNCMD :=
	RUNBIN := ./$(BINARY)
	RM := rm -f $(BINARY)
endif

BIN := $(BINARY)$(EXE)

.PHONY: all build test fmt fmt-check vet tidy clean install run example

all: build

build: fmt-check $(GOFILES)
	$(GO) build -o $(BIN) $(PKG)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

fmt-check:
	$(GO) fmt ./...

vet:
	go vet ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install $(PKG)

run: build
	$(RUNCMD) $(RUNBIN)

example:
	$(RUNCMD) $(RUNBIN) -e tbot.example.toml

clean:
	-$(RM)
