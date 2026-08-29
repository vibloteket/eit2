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
PACKAGES="$ROOT/dist/packages"
TMP="$ROOT/dist/verify-$TARGET"
rm -rf "$TMP"
mkdir -p "$TMP"

if [ "$TARGET" = linux ]; then
  name="eit2-linux-x86_64-v$VERSION"
  archive="$PACKAGES/$name.tar.gz"
  (cd "$PACKAGES" && sha256sum -c "$name.tar.gz.sha256")
  tar -xzf "$archive" -C "$TMP"
  directory="$TMP/$name"
  [ "$($directory/eit2 --version)" = "Eit 2 v$VERSION" ]
  [ -x "$directory/eit2.sh" ]
  file "$directory/eit2" | grep -q 'ELF 64-bit'
else
  name="eit2-web-v$VERSION"
  archive="$PACKAGES/$name.zip"
  (cd "$PACKAGES" && sha256sum -c "$name.zip.sha256")
  go run ./scripts/unzip "$archive" "$TMP"
  directory="$TMP/$name"
  grep -q "eit2.wasm?v=$VERSION" "$directory/index.html"
  [ -s "$directory/eit2.wasm" ]
fi

for file in LICENSE NOTICE.md ASSETS.md LICENSES/Apache-2.0.txt; do
  [ -s "$directory/$file" ]
done
printf '%s package verification passed for v%s\n' "$TARGET" "$VERSION"
