#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
VERSION=$(cat VERSION)
PACKAGES="$ROOT/dist/packages"
TMP="$ROOT/dist/verify-packages"

rm -rf "$TMP"
mkdir -p "$TMP/linux" "$TMP/windows" "$TMP/web"

(cd "$PACKAGES" && sha256sum -c SHA256SUMS)
tar -xzf "$PACKAGES/eit2-linux-x86_64-v$VERSION.tar.gz" -C "$TMP/linux"
go run ./scripts/unzip "$PACKAGES/eit2-windows-x86_64-v$VERSION.zip" "$TMP/windows"
go run ./scripts/unzip "$PACKAGES/eit2-web-v$VERSION.zip" "$TMP/web"

linux="$TMP/linux/eit2-linux-x86_64-v$VERSION"
windows="$TMP/windows/eit2-windows-x86_64-v$VERSION"
web="$TMP/web/eit2-web-v$VERSION"

[ "$($linux/eit2 --version)" = "Eit 2 v$VERSION" ]
[ -x "$linux/eit2.sh" ]
file "$linux/eit2" | grep -q 'ELF 64-bit'
file "$windows/eit2.exe" | grep -q 'PE32+ executable'
grep -q "eit2.wasm?v=$VERSION" "$web/index.html"
[ -s "$web/eit2.wasm" ]

for dir in "$linux" "$windows" "$web"; do
  [ -s "$dir/LICENSE" ]
  [ -s "$dir/NOTICE.md" ]
  [ -s "$dir/ASSETS.md" ]
  [ -s "$dir/LICENSES/Apache-2.0.txt" ]
done

printf 'Package verification passed for v%s\n' "$VERSION"
