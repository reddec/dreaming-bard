PROJECT := dreaming-bard
VERSION ?= $(shell git describe --tags --always --dirty --match=v* | sed s/v// 2> /dev/null || echo 'dev')
REPO ?= ghcr.io/reddec/$(PROJECT):$(VERSION)

# build must build and push docker image
build:
	rm -rf dist
	mkdir -p dist
	CGO_ENABLED=0 go build -ldflags='-s -w -X main.version=$(VERSION)' -trimpath -o dist/$(PROJECT) main.go
.PHONY: build

tag:
	git tag v$(shell svu --force-patch-increment --strip-prefix)
	git push --tags

# Generate code
generate:
	go generate ./...


# DEV release - local image only
docker: build
	cd dist  && docker build -f ../Dockerfile -t "$(REPO)" .

# Run locally (and build)
run:
	mkdir -p build
	go build -o build/$(PROJECT)
	./build/$(PROJECT)

.PHONY: run generate