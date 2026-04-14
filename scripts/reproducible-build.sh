#!/usr/bin/env bash
# Build Pesa with reproducible flags. Run with SOURCE_DATE_EPOCH set for deterministic timestamps.
set -euo pipefail

export CGO_ENABLED=0
export GOFLAGS="-trimpath -buildvcs=false"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
COMMIT="${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo unknown)}"
DATE="${SOURCE_DATE_EPOCH:-$(date -u +%s)}"
DATE_ISO="$(date -u -d "@${DATE}" +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -r "${DATE}" +%Y-%m-%dT%H:%M:%SZ)"

LDFLAGS="-s -w -buildid="
LDFLAGS="${LDFLAGS} -X main.version=${VERSION}"
LDFLAGS="${LDFLAGS} -X main.commit=${COMMIT}"
LDFLAGS="${LDFLAGS} -X main.date=${DATE_ISO}"

mkdir -p bin
go build -ldflags "${LDFLAGS}" -o bin/pesa ./cmd/pesa

echo "Built bin/pesa"
echo "  version: ${VERSION}"
echo "  commit:  ${COMMIT}"
echo "  date:    ${DATE_ISO}"
sha256sum bin/pesa
