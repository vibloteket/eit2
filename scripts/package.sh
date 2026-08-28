#!/bin/sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT"
VERSION=$(cat VERSION)
LDFLAGS="-s -w -X github.com/vibloteket/eit2/internal/version.Value=$VERSION"
PACKAGES="$ROOT/dist/packages"
STAGE="$ROOT/dist/stage"

rm -rf "$PACKAGES" "$STAGE"
mkdir -p "$PACKAGES" "$STAGE"

copy_legal() {
  destination=$1
  cp LICENSE NOTICE.md ASSETS.md "$destination/"
  cp -R LICENSES "$destination/"
}

linux_dir="$STAGE/eit2-linux-x86_64-v$VERSION"
mkdir -p "$linux_dir"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$LDFLAGS" -o "$linux_dir/eit2" ./cmd/eit2
cp packaging/linux/eit2.sh "$linux_dir/eit2.sh"
chmod +x "$linux_dir/eit2" "$linux_dir/eit2.sh"
cp packaging/README.txt "$linux_dir/README.txt"
copy_legal "$linux_dir"
tar -C "$STAGE" -czf "$PACKAGES/$(basename "$linux_dir").tar.gz" "$(basename "$linux_dir")"

windows_dir="$STAGE/eit2-windows-x86_64-v$VERSION"
mkdir -p "$windows_dir"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$LDFLAGS -H windowsgui" -o "$windows_dir/eit2.exe" ./cmd/eit2
cp packaging/README.txt "$windows_dir/README.txt"
copy_legal "$windows_dir"
go run ./scripts/zipdir "$windows_dir" "$PACKAGES/$(basename "$windows_dir").zip"

make build-web
web_dir="$STAGE/eit2-web-v$VERSION"
mkdir -p "$web_dir"
cp -R dist/web/. "$web_dir/"
cp packaging/README.txt "$web_dir/README.txt"
copy_legal "$web_dir"
go run ./scripts/zipdir "$web_dir" "$PACKAGES/$(basename "$web_dir").zip"

(
  cd "$PACKAGES"
  sha256sum ./* > SHA256SUMS
)

printf 'Created packages in %s\n' "$PACKAGES"
