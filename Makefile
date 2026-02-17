VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test install clean lint

build:
	go build -ldflags "$(LDFLAGS)" -o bin/share ./cmd/share

test:
	go test ./...

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/share

clean:
	rm -rf bin/

lint:
	go vet ./...
