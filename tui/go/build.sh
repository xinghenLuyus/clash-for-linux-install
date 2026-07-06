#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
GO_DIR="${ROOT_DIR}/tui/go"
OUT_DIR="${GO_DIR}/dist"
TARGET=${1:-local}

mkdir -p "${ROOT_DIR}/bin" "$OUT_DIR"
export GOCACHE="${ROOT_DIR}/.cache/go-build"
export GOMODCACHE="${ROOT_DIR}/.cache/go-mod"
mkdir -p "$GOCACHE" "$GOMODCACHE"

build_one() {
    local goos=$1 goarch=$2 out=$3
    (
        cd "$GO_DIR"
        GOOS=$goos GOARCH=$goarch CGO_ENABLED=0 go build -o "$out" ./cmd/clashctl-tui
    )
    printf 'built %s\n' "$out"
}

case "$TARGET" in
local)
    (
        cd "$GO_DIR"
        go build -o "${ROOT_DIR}/bin/clashctl-tui" ./cmd/clashctl-tui
    )
    printf 'built %s\n' "${ROOT_DIR}/bin/clashctl-tui"
    ;;
linux-amd64)
    build_one linux amd64 "${OUT_DIR}/clashctl-tui-linux-amd64"
    ;;
linux-386)
    build_one linux 386 "${OUT_DIR}/clashctl-tui-linux-386"
    ;;
linux)
    build_one linux amd64 "${OUT_DIR}/clashctl-tui-linux-amd64"
    build_one linux 386 "${OUT_DIR}/clashctl-tui-linux-386"
    ;;
all)
    "$0" local
    "$0" linux
    ;;
*)
    printf 'usage: %s [local|linux-amd64|linux-386|linux|all]\n' "$0" >&2
    exit 1
    ;;
esac
