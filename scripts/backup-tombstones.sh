#!/usr/bin/env bash
# backup-tombstones.sh <backup_id> <staging_dir>
# Invoked by mariabackup.sh between --prepare and tar+gpg to strip
# tombstoned users (PRIV-04). Exits non-zero on any failure so the caller
# aborts rather than shipping an unstripped backup.
set -euo pipefail

BACKUP_ID="${1:?backup_id required}"
STAGING_DIR="${2:?staging_dir required}"

log() { printf '[%s] backup-tombstones %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*"; }

log "start backup_id=$BACKUP_ID staging=$STAGING_DIR"
if [ ! -d "$STAGING_DIR" ]; then
  log "staging dir missing — aborting"
  exit 1
fi

PESA_BIN="${PESA_BIN:-pesa}"

if ! command -v "$PESA_BIN" >/dev/null 2>&1; then
  log "pesa binary not found on PATH (PESA_BIN=$PESA_BIN)"
  exit 1
fi

"$PESA_BIN" apply-tombstones "$BACKUP_ID" --staging-dir "$STAGING_DIR"

log "done backup_id=$BACKUP_ID"
