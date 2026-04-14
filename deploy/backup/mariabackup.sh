#!/usr/bin/env bash
# mariabackup.sh {full|incremental|--check}
# Encrypts with GPG and rsyncs to a Hetzner Storage Box.
set -euo pipefail

MODE="${1:-full}"
DATE="$(date -u +%Y-%m-%dT%H%M%SZ)"
BASE="/var/lib/pesa-backup"
STAGE="$BASE/stage/$DATE"
ENC="$BASE/enc"
LAST_FULL_LINK="$BASE/last-full"
: "${BACKUP_GPG_RECIPIENT:?BACKUP_GPG_RECIPIENT is required}"
: "${HETZNER_RSYNC_TARGET:?HETZNER_RSYNC_TARGET is required (e.g. u1@u1.your-storagebox.de:/home/pesa/backups)}"
: "${MARIABACKUP_USER:=pesa_backup}"
: "${MARIABACKUP_PASSWORD:?MARIABACKUP_PASSWORD is required}"

mkdir -p "$STAGE" "$ENC"

case "$MODE" in
  --check)
    gpg --list-keys "$BACKUP_GPG_RECIPIENT" >/dev/null
    ssh -o BatchMode=yes -o ConnectTimeout=10 "${HETZNER_RSYNC_TARGET%%:*}" true
    echo "check OK: GPG recipient present, Hetzner reachable"
    exit 0 ;;
  full)
    mariabackup --backup --target-dir="$STAGE" \
      --user="$MARIABACKUP_USER" --password="$MARIABACKUP_PASSWORD" --ssl
    mariabackup --prepare --target-dir="$STAGE"
    rm -f "$LAST_FULL_LINK"
    ln -s "$STAGE" "$LAST_FULL_LINK"
    ;;
  incremental)
    test -L "$LAST_FULL_LINK" || { echo "no full backup yet"; exit 1; }
    mariabackup --backup --target-dir="$STAGE" \
      --incremental-basedir="$(readlink -f "$LAST_FULL_LINK")" \
      --user="$MARIABACKUP_USER" --password="$MARIABACKUP_PASSWORD" --ssl
    ;;
  *) echo "usage: $0 {full|incremental|--check}"; exit 2 ;;
esac

# PRIV-04: strip tombstoned users from the prepared staging snapshot BEFORE
# tar+gpg. Any failure aborts the backup (better to retry than ship data).
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [ -x "$SCRIPT_DIR/../../scripts/backup-tombstones.sh" ]; then
  "$SCRIPT_DIR/../../scripts/backup-tombstones.sh" "$DATE" "$STAGE"
elif [ -x "scripts/backup-tombstones.sh" ]; then
  scripts/backup-tombstones.sh "$DATE" "$STAGE"
else
  echo "backup-tombstones.sh missing — aborting" >&2
  exit 1
fi

TAR="$BASE/${MODE}-${DATE}.tar.zst"
tar --zstd -cf "$TAR" -C "$STAGE" .
gpg --yes --batch --trust-model always --recipient "$BACKUP_GPG_RECIPIENT" \
    --encrypt --output "$ENC/$(basename "$TAR").gpg" "$TAR"
rm -f "$TAR"
rsync -a --remove-source-files "$ENC/" "$HETZNER_RSYNC_TARGET/"
# retention: keep last 14 stages locally
find "$BASE/stage" -maxdepth 1 -mindepth 1 -mtime +14 -exec rm -rf {} \;
echo "backup $MODE $DATE OK"
