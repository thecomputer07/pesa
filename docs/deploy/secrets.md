# Environment variable dictionary

Generate sensitive values with `openssl rand` or the listed command. Write them
to `/etc/pesa/env` with `umask 077`, owner `pesa:pesa`, mode `0600`.

| Name | Purpose | Generate / Source |
|------|---------|-------------------|
| `PESA_ENV` | `prod` enables strict checks (TLS, LOG_SALT, etc.) | literal `prod` |
| `PESA_ADDR` | bind address (default `:8080`) | literal |
| `PESA_ENABLE_TLS` | `1` enables autocert + TLS 1.3 on `:443` | literal |
| `DATABASE_DSN` | `pesa_app:PW@tcp(127.0.0.1:3306)/pesa?parseTime=true&tls=true` | composed |
| `LOG_SALT` | pseudonymise email in logs (PRIV-06) | `openssl rand -hex 32` |
| `COOKIE_DOMAIN` | matches `pesadia.com` | literal |
| `AUTOCERT_HOSTS` | comma-separated FQDNs | literal |
| `AUTOCERT_EMAIL` | Let's Encrypt account email | `lorenzo.dimare@icloud.com` |
| `AUTOCERT_CACHE_DIR` | directory for cached certs | `/var/cache/pesa-autocert` |
| `SMTP_HOST` | e.g. `smtp.eu.mailgun.org` | provider |
| `SMTP_PORT` | `587` | literal |
| `SMTP_USER` | SMTP login | provider |
| `SMTP_PASSWORD` | SMTP password | provider |
| `SMTP_FROM` | `no-reply@pesadia.com` | literal |
| `PESA_ERASURE_HASH_SALT` | salts the erasure_log row hashes (PRIV-03) | `openssl rand -hex 32` |
| `BACKUP_GPG_RECIPIENT` | GPG key uid for backups | `backup@pesadia.com` (see `deploy/backup/README.md`) |
| `HETZNER_RSYNC_TARGET` | Hetzner Storage Box rsync URL | `u1@u1.your-storagebox.de:/home/pesa/backups` |
| `MARIABACKUP_PASSWORD` | password for `pesa_backup` SQL user | `openssl rand -base64 32` |
| `PESA_ENABLE_PPROF` | `1` mounts /debug/pprof (admin only) | unset in prod normally |
| `PPROF_TOKEN` | required header value when pprof enabled | `openssl rand -hex 32` |
