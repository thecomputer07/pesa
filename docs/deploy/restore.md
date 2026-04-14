# Restore playbook

Target SLA: < 4h restoration from cold (SEC-36).

## 1. Pull encrypted archives

```bash
mkdir -p /tmp/restore && cd /tmp/restore
rsync -a "$HETZNER_RSYNC_TARGET/" .
```

## 2. Decrypt

```bash
for f in *.tar.zst.gpg; do
  gpg --decrypt --output "${f%.gpg}" "$f"
done
```

## 3. Extract

```bash
mkdir -p full inc
tar --zstd -xf full-*.tar.zst -C full
for inc_archive in $(ls -1 incremental-*.tar.zst | sort); do
  d="inc/$(basename "$inc_archive" .tar.zst)"
  mkdir -p "$d" && tar --zstd -xf "$inc_archive" -C "$d"
done
```

## 4. Prepare

```bash
mariabackup --prepare --target-dir=full
for d in $(ls -1d inc/* | sort); do
  mariabackup --prepare --target-dir=full --incremental-dir="$d"
done
```

## 5. Stop, copy back, restart

```bash
systemctl stop mariadb     # or container
mariabackup --copy-back --target-dir=full --datadir=/var/lib/mysql
chown -R mysql:mysql /var/lib/mysql
systemctl start mariadb
```

## 6. Replay app migrations + restart

```bash
sudo -u pesa /opt/pesa/bin/pesa migrate up
systemctl restart pesa
```

## 7. Verify

- Login with a known test user.
- `curl -sf https://pesadia.com/readyz` returns `{"status":"ok"}`.
- Spot-check a recent meal entry.

## 8. Drill cadence

Run end-to-end every quarter on a clean VPS. Log in `docs/deploy/restore-drill-log.md`.
Drift > 4h triggers an action plan to reduce backup size or accelerate decrypt/copy.
