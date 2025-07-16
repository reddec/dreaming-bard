PROJECT := dreaming-bard
VERSION ?= $(shell git describe --tags --always --dirty --match=v* | sed s/v// 2> /dev/null || echo 'dev')
REPO ?= ghcr.io/reddec/$(PROJECT):$(VERSION)

# Generate code
generate:
	go generate ./...

# DEV release - local image only
snapshot: build
	goreleaser release --clean --snapshot

# Run locally (and build)
run:
	mkdir -p build
	go build -o build/$(PROJECT)
	./build/$(PROJECT)

.PHONY: run generate