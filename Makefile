#
#  Makefile for Go
#
SHELL=/usr/bin/env bash
VERSION=$(shell git describe --tags --always)
PACKAGES = $(shell find ./ -type d | grep -v 'vendor' | grep -v '.git' | grep -v 'bin')

default: build

.PHONY: gox
gox:
	go get -u github.com/mitchellh/gox

.PHONY: build
build:
	go build -ldflags="-X main.Version=${VERSION}" -o burry-${VERSION}

.PHONY: clean
clean:
	rm -f coverage-all.out
	rm -f coverage.out

.PHONY: binaries
binaries:
	gox github.com/mhausenblas/burry.sh
