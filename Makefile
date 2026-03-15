-include .env
export

BINARY := linkding
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build run clean snapshot release

build:
	go build -ldflags "-X github.com/chickenzord/linkding-cli/internal/version.Version=$(VERSION)" -o $(BINARY) ./cmd/linkding

run: build
	./$(BINARY) $(ARGS)

clean:
	rm -f $(BINARY)
	rm -rf dist/

snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean
