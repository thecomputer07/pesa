#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."
go build -o bin/check-forbidden-words ./cmd/check-forbidden-words

expect_exit() {
    local expected=$1 ; shift
    "$@" && rc=0 || rc=$?
    [ "$rc" -eq "$expected" ] || { echo "FAIL: expected $expected got $rc: $*"; exit 1; }
    echo "OK ($rc): $*"
}

expect_exit 1 bin/check-forbidden-words --scan-file scripts/test/anti-word-fixtures/positive-it.md
expect_exit 1 bin/check-forbidden-words --scan-file scripts/test/anti-word-fixtures/positive-en.md
expect_exit 0 bin/check-forbidden-words --scan-file scripts/test/anti-word-fixtures/negative.md

out="$(bin/check-forbidden-words --scan-file scripts/test/anti-word-fixtures/allowlist.md || true)"
echo "$out" | grep -q '^ALLOWLISTED' || { echo "FAIL: expected ALLOWLISTED in stdout"; echo "$out"; exit 1; }
echo "OK: allowlist"

err="$(bin/check-forbidden-words --scan-file scripts/test/anti-word-fixtures/soft-only.md 2>&1 1>/dev/null || true)"
echo "$err" | grep -q 'marketing-rules warning' || { echo "FAIL: expected SOFT warning"; echo "$err"; exit 1; }
echo "OK: soft warning"

echo "ALL FIXTURE TESTS PASSED"
