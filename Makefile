.PHONY: install run test lint music build build-web package package-linux package-web verify-packages verify-linux verify-web check clean

install:
	go mod download

run:
	go run ./cmd/eit2

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './dist/*'))" || (gofmt -d $$(gofmt -l $$(find . -name '*.go' -not -path './dist/*')); exit 1)

music:
	bun scripts/music/render-badinerie.ts
	bun scripts/music/render-beethoven.ts
	bun scripts/music/render-beethoven-g2.ts

VERSION := $(shell cat VERSION)
LDFLAGS := -X github.com/vibloteket/eit2/internal/version.Value=$(VERSION)

build:
	mkdir -p dist/native
	go build -trimpath -ldflags "$(LDFLAGS)" -o dist/native/eit2 ./cmd/eit2

build-web:
	mkdir -p dist/web
	GOOS=js GOARCH=wasm go build -trimpath -ldflags "$(LDFLAGS)" -o dist/web/eit2.wasm ./cmd/eit2
	sed 's/__VERSION__/$(VERSION)/g' web/index.html > dist/web/index.html
	cp web/favicon.svg web/favicon-32.png dist/web/
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/web/wasm_exec.js
	cp LICENSE NOTICE.md ASSETS.md CREDITS.md MUSIC-SOURCES.md dist/web/
	rm -rf dist/web/LICENSES
	cp -R LICENSES dist/web/

package:
	./scripts/package.sh

package-linux:
	./scripts/package-platform.sh linux

package-web:
	./scripts/package-platform.sh web

verify-packages: package
	./scripts/verify-packages.sh

verify-linux: package-linux
	./scripts/verify-platform-package.sh linux

verify-web: package-web
	./scripts/verify-platform-package.sh web

check: lint test build build-web

clean:
	rm -rf dist
