.PHONY: install run test lint build build-web check clean

install:
	go mod download

run:
	go run ./cmd/eit2

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './dist/*'))" || (gofmt -d $$(gofmt -l $$(find . -name '*.go' -not -path './dist/*')); exit 1)

build:
	mkdir -p dist/native
	go build -trimpath -o dist/native/eit2 ./cmd/eit2

build-web:
	mkdir -p dist/web
	GOOS=js GOARCH=wasm go build -trimpath -o dist/web/eit2.wasm ./cmd/eit2
	cp web/index.html web/favicon.svg web/favicon-32.png dist/web/
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/web/wasm_exec.js

check: lint test build build-web

clean:
	rm -rf dist
