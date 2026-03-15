-include .env
export

.PHONY: build run

build:
	go build -o linkding ./cmd/linkding

run: build
	./linkding $(ARGS)
