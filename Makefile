.PHONY: wire build test

wire:
	wire ./...

build:
	go build ./...

test:
	go test ./...
