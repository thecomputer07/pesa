#!/usr/bin/env bash
# Thin wrapper for local dev: builds the Go scanner if needed and runs it.
set -euo pipefail
cd "$(dirname "$0")/.."
if [ ! -x bin/check-forbidden-words ] || [ cmd/check-forbidden-words/main.go -nt bin/check-forbidden-words ]; then
  mkdir -p bin
  go build -o bin/check-forbidden-words ./cmd/check-forbidden-words
fi
exec bin/check-forbidden-words "$@"
