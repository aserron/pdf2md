# Cross-Platform Makefile for pdf2md (Windows & Unix)

BINARY_CLI=bin/pdf2md
BINARY_GUI=bin/pdf2md-gui

ifeq ($(OS),Windows_NT)
	BINARY_CLI=bin/pdf2md.exe
	BINARY_GUI=bin/pdf2md-gui.exe
	MKDIR=powershell -NoProfile -Command "if (!(Test-Path bin)) { New-Item -ItemType Directory -Path bin }"
	RM=powershell -NoProfile -Command "if (Test-Path bin) { Remove-Item -Recurse -Force bin }"
else
	MKDIR=mkdir -p bin
	RM=rm -rf bin
endif

.PHONY: all build build-cli build-gui test bench clean run-cli run-gui help

all: build

help:
	@echo "Available targets:"
	@echo "  make build       - Build both CLI and Windows GUI binaries"
	@echo "  make build-cli   - Build CLI binary"
	@echo "  make build-gui   - Build GUI binary"
	@echo "  make test        - Run automated tests with race detection"
	@echo "  make bench       - Run performance benchmarks"
	@echo "  make clean       - Remove built binaries"
	@echo "  make run-cli     - Run CLI with --help"
	@echo "  make run-gui     - Run GUI application"

build: build-cli build-gui

build-cli:
	@$(MKDIR)
	go build -trimpath -ldflags="-s -w" -o $(BINARY_CLI) ./cmd/pdf2md

build-gui:
	@$(MKDIR)
	go build -trimpath -ldflags="-s -w -H=windowsgui" -o $(BINARY_GUI) ./cmd/pdf2md-gui

test:
	go run ./cmd/testrunner

bench:
	go test -bench=. -benchmem ./...

clean:
	@$(RM)

run-cli:
	go run ./cmd/pdf2md --help

run-gui:
	go run ./cmd/pdf2md-gui
