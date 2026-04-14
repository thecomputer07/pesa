# Pesa — Production deploy runbook (OVH Milano)

Target environment: Debian 12 minimal, single VPS, Italian jurisdiction (OVH Milano).

## 1. Provision

- Order an OVH VPS in Milano (cheapest tier with 4 GB RAM is enough for MVP).
- Image: Debian 12 minimal.
- Boot, set hostname, configure `/etc/hosts`, `timedatectl set-timezone Europe/Rome`.

## 2. First-boot hardening

```bash
apt update && apt full-upgrade -y
adduser --disabled-password --gecos '' pesa
usermod -aG sudo pesa
mkdir -p /home/pesa/.ssh && cp ~/.ssh/authorized_keys /home/pesa/.ssh/
chown -R pesa:pesa /home/pesa/.ssh && chmod 700 /home/pesa/.ssh

# SSH: disable password auth, change port, root login off
sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config
systemctl restart ssh

apt install -y unattended-upgrades fail2ban ufw mariadb-backup rsync gnupg curl
dpkg-reconfigure -plow unattended-upgrades
systemctl enable --now fail2ban
```

## 3. Firewall

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 22/tcp        # or your custom SSH port
ufw enable
```

## 4. Docker stack

```bash
apt install -y docker.io docker-compose-plugin
usermod -aG docker pesa
mkdir -p /opt/pesa/deploy/secrets
chown -R pesa:pesa /opt/pesa
# Place secret files (each containing exactly the password / key value, no trailing newline ideally):
#   deploy/secrets/{mariadb_root,mariadb_app,plausible_db,plausible_db_url,plausible_secret,glitchtip_db,glitchtip_db_url,glitchtip_secret}
chmod 600 /opt/pesa/deploy/secrets/*
sudo -u pesa docker compose -f /opt/pesa/deploy/docker-compose.yml up -d
```

## 5. Application

```bash
install -o pesa -g pesa -m 0755 bin/pesa /opt/pesa/bin/pesa
install -d -o pesa -g pesa -m 0750 /var/lib/pesa /var/cache/pesa-autocert
install -d -m 0750 /etc/pesa
install -m 0600 /dev/stdin /etc/pesa/env <<'EOF'
PESA_ENV=prod
PESA_ENABLE_TLS=1
PESA_ADDR=:8080
DATABASE_DSN=pesa_app:CHANGE@tcp(127.0.0.1:3306)/pesa?parseTime=true&tls=true
LOG_SALT=CHANGE
COOKIE_DOMAIN=pesadia.com
AUTOCERT_HOSTS=pesadia.com,www.pesadia.com
AUTOCERT_EMAIL=lorenzo.dimare@icloud.com
AUTOCERT_CACHE_DIR=/var/cache/pesa-autocert
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=no-reply@pesadia.com
PESA_ERASURE_HASH_SALT=CHANGE
EOF
chown pesa:pesa /etc/pesa/env

sudo -u pesa /opt/pesa/bin/pesa migrate up
install -m 0644 deploy/systemd/pesa.service /etc/systemd/system/pesa.service
systemctl daemon-reload
systemctl enable --now pesa
```

## 6. DNS + TLS

- Create A records for `pesadia.com`, `www.pesadia.com`, `analytics.pesadia.com`, `errors.pesadia.com` pointing to the VPS IPv4.
- Add a CAA record: `0 issue "letsencrypt.org"`.
- autocert provisions Let's Encrypt certs on first request.

## 7. Backups

```bash
install -m 0755 deploy/backup/mariabackup.sh /opt/pesa/deploy/backup/mariabackup.sh
install -m 0644 deploy/cron/pesa-backup.cron /etc/cron.d/pesa-backup
```

Configure GPG key + Hetzner Storage Box per `deploy/backup/README.md`. Then:

```bash
BACKUP_GPG_RECIPIENT=backup@pesadia.com \
HETZNER_RSYNC_TARGET=u1234567@u1234567.your-storagebox.de:/home/pesa/backups \
MARIABACKUP_PASSWORD=CHANGE \
  /opt/pesa/deploy/backup/mariabackup.sh --check
```

## 8. Smoke

```bash
curl -sf https://pesadia.com/healthz
curl -sf https://pesadia.com/readyz
```

## 9. Next reads

- `ssl-labs-checklist.md` — A+ verification.
- `restore.md` — quarterly drill.
- `secrets.md` — env var dictionary.
