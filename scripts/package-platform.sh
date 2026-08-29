#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || { [ "$1" != linux ] && [ "$1" != web ]; }; then
  echo "usage: $0 linux|web" >&2
  exit 2
fi

TARGET=$1
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
VERSION=$(cat VERSION)
LDFLAGS="-s -w -X github.com/vibloteket/eit2/internal/version.Value=$VERSION"
PACKAGES="$ROOT/dist/packages"
STAGE="$ROOT/dist/stage-$TARGET"
rm -rf "$PACKAGES" "$STAGE"
mkdir -p "$PACKAGES" "$STAGE"

copy_legal() {
  destination=$1
  cp LICENSE NOTICE.md ASSETS.md "$destination/"
  cp -R LICENSES "$destination/"
}

if [ "$TARGET" = linux ]; then
  package="eit2-linux-x86_64-v$VERSION.tar.gz"
  directory="$STAGE/eit2-linux-x86_64-v$VERSION"
  mkdir -p "$directory"
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$LDFLAGS" -o "$directory/eit2" ./cmd/eit2
  cp packaging/linux/eit2.sh "$directory/eit2.sh"
  chmod +x "$directory/eit2" "$directory/eit2.sh"
  cp packaging/README.txt "$directory/README.txt"
  copy_legal "$directory"
  tar -C "$STAGE" -czf "$PACKAGES/$package" "$(basename "$directory")"
else
  package="eit2-web-v$VERSION.zip"
  directory="$STAGE/eit2-web-v$VERSION"
  make build-web
  mkdir -p "$directory"
  cp -R dist/web/. "$directory/"
  cp packaging/README.txt "$directory/README.txt"
  copy_legal "$directory"
  go run ./scripts/zipdir "$directory" "$PACKAGES/$package"
fi

(cd "$PACKAGES" && sha256sum "$package" > "$package.sha256")
printf 'Created %s\n' "$PACKAGES/$package"
