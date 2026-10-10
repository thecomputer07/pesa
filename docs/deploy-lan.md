# Deploy Pesa sulla LAN (es. http://10.0.0.2:8080)

Questo runbook mostra come esporre Pesa su una rete locale — tipicamente per testare la PWA da cellulare o tablet, o per condividere l'istanza con persone sullo stesso WiFi.

## Prerequisiti

- Docker Desktop (o Docker Engine + compose)
- IP statico o stabile della macchina host (scopri con `ipconfig` su Windows)
- Porta **8080** libera sull'host

## Procedura

### 1. Ferma il compose di sviluppo (se attivo)

Il compose di sviluppo e quello LAN usano entrambi la porta 8080 — non possono girare contemporaneamente.

```bash
docker compose down
```

### 2. Opzionale: personalizza i secrets

```bash
cp .env.lan.example .env.lan
# Edita .env.lan con i tuoi secret (openssl rand -hex 32).
# Se salti questo passo, il compose usa default insicuri (OK per test interno).
```

### 3. Build + avvio

```bash
# Con .env.lan personalizzato:
docker compose --env-file .env.lan -f docker-compose.lan.yml up -d --build

# Oppure con i default:
docker compose -f docker-compose.lan.yml up -d --build
```

La prima build impiega 2–4 minuti (npm ci + vite build + go build). Dopo il warm cache ~30 s.

### 4. Apri Windows Firewall (una sola volta)

Su Windows, il firewall blocca connessioni in ingresso sulla porta 8080. Apri **PowerShell come amministratore** ed esegui:

```powershell
New-NetFirewallRule -DisplayName "Pesa LAN 8080" -Direction Inbound -Protocol TCP -LocalPort 8080 -Action Allow
```

Per rimuoverla:

```powershell
Remove-NetFirewallRule -DisplayName "Pesa LAN 8080"
```

Su macOS/Linux di solito non serve (firewall di default permissivo sulla LAN).

### 5. Verifica

**Dalla stessa macchina:**
```
http://localhost:8080
```

**Da un altro device sulla LAN (telefono, portatile):**
```
http://10.0.0.2:8080
```

(Sostituisci `10.0.0.2` con l'IP LAN del tuo host — vedi `ipconfig`.)

### 6. Logs / troubleshooting

```bash
docker compose -f docker-compose.lan.yml logs -f backend
docker compose -f docker-compose.lan.yml logs -f db
```

### 7. Stop

```bash
docker compose -f docker-compose.lan.yml down

# Per cancellare anche il DB:
docker compose -f docker-compose.lan.yml down -v
```

## Avvertenze

### PWA install e service worker

HTTP (non HTTPS) → il service worker PWA **funziona solo su `localhost`** e **sui suoi IP LAN privati** (come `10.0.0.2`) nei browser moderni, con limitazioni. In Chrome/Edge i PWA su HTTP privato sono installabili. In Safari iOS l'install è più restrittivo — potrebbe richiedere HTTPS vero.

### Cookie e sessioni

`COOKIE_SECURE=false` per HTTP. I cookie di sessione restano `HttpOnly` + `SameSite=Lax`. Non usare questo setup per dati reali — è solo per test/beta interni.

### Email di verifica

Nel LAN deployment SMTP è disabilitato (`SMTP_DISABLED=true`). I link di verifica vengono loggati in console backend invece che spediti. Recuperali con:

```bash
docker compose -f docker-compose.lan.yml logs backend | grep "verify"
```

### HTTPS

Questa configurazione è HTTP. Per HTTPS sulla LAN ci sono due strade:
1. **mkcert** + reverse proxy (Caddy/nginx/Traefik) con certificati self-signed
2. **Autocert (Let's Encrypt)** — richiede un dominio pubblico puntato alla tua IP e porta 80/443 aperte. Non è il caso della LAN pura.

Per beta interno HTTP basta.

## Quando stai in produzione vera

Questa config NON è production-ready. Per il launch su pesadia.com:

- Usa `PESA_ENV=prod` (abilita check TLS, richiede `sslmode=require`)
- Genera tutti i secrets con `openssl rand -hex 32`
- `PESA_ENABLE_TLS=1` + `AUTOCERT_HOSTS=pesadia.com` + `AUTOCERT_EMAIL=...`
- `COOKIE_SECURE=true`, `COOKIE_DOMAIN=pesadia.com`
- Firewall: solo 80/443 esposte, DB isolato
- Backup periodici configurati
- SMTP reale (es. Amazon SES, Postmark)
