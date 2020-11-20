SHELL := /bin/bash

# Default hyperdark version number to the shorthand git commit hash if
# not set at the command line.
VER     := $(or $(VER),$(shell git log -1 --format="%h"))
COMMIT  := $(shell git log -1 --format="%h - %ae")
DATE    := $(shell date -u)
VERSION := $(VER) (commit $(COMMIT)) $(DATE)

GOSOURCES := $(shell find . \( -name '*.go' \))
TEMPLATES := $(shell find tmpl/templates \( -name '*' \))

THISFILE := $(lastword $(MAKEFILE_LIST))
THISDIR  := $(shell dirname $(realpath $(THISFILE)))
GOBIN    := $(THISDIR)/bin

# Prepend this repo's bin directory to our path since we'll want to
# install some build tools there for use during the build process.
PATH := $(GOBIN):$(PATH)

# Export GOBIN env variable so `go install` picks it up correctly.
export GOBIN

all: 	install-build-deps bin/ccDropper	
clean:
	-rm -rf bin

.PHONY: install-build-deps
install-build-deps: bin/go-bindata
	go get gopkg.in/yaml.v3

.PHONY: remove-build-deps
remove-build-deps:
	$(RM) bin/go-bindata

bin/go-bindata:
	go get github.com/go-bindata/go-bindata/go-bindata

tmpl/bindata.go: $(TEMPLATES) bin/go-bindata
	$(GOBIN)/go-bindata -pkg tmpl -prefix tmpl/templates -o tmpl/bindata.go tmpl/templates/...

bin/ccDropper: $(GOSOURCES) tmpl/bindata.go
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X 'phenix/version.Version=$(VERSION)' -s -w" -trimpath -o bin/phenix-app-ccDropper src/ccDropper.go src/phenixDefs.go
