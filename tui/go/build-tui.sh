#!/usr/bin/env bash
set -euo pipefail

MODULE_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
ROOT="$(cd -- "${MODULE_DIR}/../.." && pwd -P)"
OUT="${ROOT}/bin/clashctl-tui"
GO_BIN="${GO_BIN:-}"
export GOCACHE="${GOCACHE:-${ROOT}/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-${ROOT}/.cache/go-mod}"
TARGET="${1:-local}"

if [ -z "$GO_BIN" ]; then
    GO_BIN="$(command -v go 2>/dev/null || true)"
fi
if [ -z "$GO_BIN" ] && [ -x /usr/local/go/bin/go ]; then
    GO_BIN=/usr/local/go/bin/go
fi
if [ -z "$GO_BIN" ]; then
    printf 'go not found. Install Go or run with Docker, for example:\n' >&2
    printf '  docker run --rm -v "$PWD":/work -w /work golang:1.22 bash -lc "cd tui/go && /usr/local/go/bin/go build -o ../../bin/clashctl-tui ./cmd/clashctl-tui"\n' >&2
    exit 1
fi

mkdir -p "${ROOT}/bin" "$GOCACHE" "$GOMODCACHE"

build_one() {
    local goos=$1 goarch=$2 out=$3
    if [ "$goos" = local ]; then
        (cd "$MODULE_DIR" && CGO_ENABLED=0 "$GO_BIN" build -o "$out" ./cmd/clashctl-tui)
    else
        (cd "$MODULE_DIR" && GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 "$GO_BIN" build -o "$out" ./cmd/clashctl-tui)
    fi
    printf 'built %s\n' "$out"
}

case "$TARGET" in
local)
    build_one local local "$OUT"
    ;;
linux-amd64 | amd64 | x86_64)
    build_one linux amd64 "${ROOT}/bin/clashctl-tui-linux-amd64"
    ;;
linux-386 | 386 | x86)
    build_one linux 386 "${ROOT}/bin/clashctl-tui-linux-386"
    ;;
linux)
    build_one linux amd64 "${ROOT}/bin/clashctl-tui-linux-amd64"
    build_one linux 386 "${ROOT}/bin/clashctl-tui-linux-386"
    ;;
all)
    build_one local local "$OUT"
    build_one linux amd64 "${ROOT}/bin/clashctl-tui-linux-amd64"
    build_one linux 386 "${ROOT}/bin/clashctl-tui-linux-386"
    ;;
-h | --help)
    cat <<EOF
Usage:
  tui/go/build-tui.sh [local|linux|linux-amd64|linux-386|all]

Outputs:
  local        bin/clashctl-tui
  linux-amd64  bin/clashctl-tui-linux-amd64
  linux-386    bin/clashctl-tui-linux-386
EOF
    ;;
*)
    printf 'unknown target: %s\n' "$TARGET" >&2
    printf 'run: tui/go/build-tui.sh --help\n' >&2
    exit 1
    ;;
esac
