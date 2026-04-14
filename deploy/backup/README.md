# Backup operator README

This directory ships `mariabackup.sh` and `pesa-backup.cron`.

## Required GPG key

The script encrypts every archive to `BACKUP_GPG_RECIPIENT`. Generate the key
once on a secure workstation (NOT on the server):

```bash
gpg --batch --full-generate-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Name-Real: Pesa Backup
Name-Email: backup@pesadia.com
Expire-Date: 2y
%commit
EOF
gpg --armor --export backup@pesadia.com > pesa-backup.pub.asc
```

Import the **public key only** on the production server:

```bash
gpg --import pesa-backup.pub.asc
gpg --edit-key backup@pesadia.com   # set trust to ultimate
```

Keep the private key offline (USB stick, encrypted vault). It is required to
**restore** backups (see `docs/deploy/restore.md`).

## Required Hetzner Storage Box

Provision via Hetzner Robot panel. Enable SSH key auth. Add the server's
public key to `~/.ssh/authorized_keys` on the box.

## Verification

```bash
BACKUP_GPG_RECIPIENT=backup@pesadia.com \
HETZNER_RSYNC_TARGET=u1234567@u1234567.your-storagebox.de:/home/pesa/backups \
  ./mariabackup.sh --check
```
