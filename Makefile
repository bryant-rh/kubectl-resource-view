
PKGROOT=github.com/bryant-rh/kubectl-resource-view

CMD=kubectl-resource-view
VERSION ?= $(shell git describe --abbrev=0 --tags 2>/dev/null || echo no-version)
GIT_COMMIT := $(shell git rev-parse --short HEAD)

OUTDIR=_output

GOLDFLAGS=-w -X $(PKGROOT)/cmd/kubectl-resource-view.version=$(VERSION)

export GO111MODULE=on

.PHONY: build build-linux build-darwin build-windows install check clean

build:
	CGO_ENABLED=0 \
	go build -ldflags "$(GOLDFLAGS)" -o $(OUTDIR)/$(CMD)

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags "$(GOLDFLAGS)" -o $(OUTDIR)/$(CMD)_linux-amd64/$(CMD)
	
build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
	go build -ldflags "$(GOLDFLAGS)" -o $(OUTDIR)/$(CMD)_linux-arm64/$(CMD)

build-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags "$(GOLDFLAGS)" -o $(OUTDIR)/$(CMD)_darwin-amd64/$(CMD)

build-windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
	go build -ldflags "$(GOLDFLAGS)" -o $(OUTDIR)/$(CMD)_windows-amd64/$(CMD).exe

install:
	CGO_ENABLED=0 go install -ldflags "$(GOLDFLAGS)"

clean:
	-rm -r $(OUTDIR)