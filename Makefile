.PHONY: run build test fmt

GOCACHE ?= /tmp/go-build
GOMODCACHE ?= /tmp/go-mod

run:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go run ./cmd/admin-panel

build:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build -o bin/admin-panel ./cmd/admin-panel

test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./...

fmt:
	gofmt -w ./cmd ./internal
