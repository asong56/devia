#!/bin/bash
# Builds devia for Linux/Windows/macOS x amd64/arm64, in two variants:
#   devia-<os>-<arch>       full build: CLI + `devia serve` (JSON API)
#   devia-cli-<os>-<arch>   -tags noserve: CLI only, no net/http linked
#
# Every build is -trimpath -ldflags="-s -w" (strip symbol table + DWARF
# debug info + embedded file paths). This alone typically cuts a Go
# binary's size by 25-35%.
set -euo pipefail

mkdir -p build
LDFLAGS="-s -w"

platforms=(
  "linux amd64"
  "linux arm64"
  "windows amd64"
  "windows arm64"
  "darwin amd64"
  "darwin arm64"
)

for p in "${platforms[@]}"; do
  set -- $p
  os=$1; arch=$2
  ext=""
  [ "$os" = "windows" ] && ext=".exe"

  echo "==> $os/$arch (full: CLI + serve)"
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
    go build -trimpath -ldflags="$LDFLAGS" -o "build/devia-$os-$arch$ext" .

  echo "==> $os/$arch (cli-only, -tags noserve)"
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
    go build -tags noserve -trimpath -ldflags="$LDFLAGS" -o "build/devia-cli-$os-$arch$ext" .
done

echo ""
echo "Done:"
ls -lh build/
