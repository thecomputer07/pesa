# Setup ambiente di sviluppo locale

Questa guida ti porta da zero a "app gira sulla mia macchina" in ~10 minuti.

## Requisiti

| Tool | Versione | Verifica |
|---|---|---|
| Go | 1.22+ | `go version` |
| Node.js | 20+ | `node --version` |
| Docker Desktop | recente | `docker --version` |
| make | qualsiasi | `make --version` (su Windows: WSL o `choco install make`) |

## 1. Clona / posizionati nel repo

```bash
cd C:/Users/TheComputer07/desktop/diabetic
```

## 2. Avvia MariaDB locale

```bash
docker compose -f docker-compose.dev.yml up -d
```

Verifica che sia healthy:

```bash
docker compose -f docker-compose.dev.yml ps
```

UI web del DB su http://localhost:8081 (Adminer — login: server `mariadb`, user `pesa`, password `pesa`, db `pesa`).

## 3. Carica le variabili d'ambiente

```bash
# Linux/macOS/WSL
export $(grep -v '^#' .env.dev | xargs)

# Windows PowerShell
Get-Content .env.dev | ForEach-Object {
  if ($_ -match '^([^#=]+)=(.*)$') { [System.Environment]::SetEnvironmentVariable($Matches[1], $Matches[2]) }
}

# Windows Git Bash
set -a; source .env.dev; set +a
```

Oppure usa `direnv` se installato (`direnv allow`).

## 4. Installa dipendenze

```bash
go mod tidy
cd web && npm install && cd ..
```

## 5. Build + migrazioni

```bash
make build
./bin/pesa migrate up
```

Output atteso: `OK applied 24 migrations` (o simile).

## 6. Avvia (due terminali)

**Terminale 1 — backend Go:**
```bash
./bin/pesa serve
# o per hot-reload: make dev
```

**Terminale 2 — frontend Vite:**
```bash
cd web && npm run dev
```

## 7. Apri nel browser

http://localhost:5173

- Homepage con scope banner
- `/auth/register` — crea account (link verifica email appare nei log Go invece di mandare mail)
- `/cerca` — ricerca alimenti (vuota finché non popoli il DB)

## 8. Popola dati di prova

```bash
docker compose -f docker-compose.dev.yml exec mariadb \
  mariadb -upesa -ppesa pesa <<'SQL'
INSERT INTO foods (source_id, name_it, kcal, protein_g, fat_g, carb_g)
VALUES (1, 'Pasta di semola', 357, 12.5, 1.4, 73.6),
       (1, 'Mela', 52, 0.3, 0.2, 14.0),
       (1, 'Parmigiano Reggiano', 392, 33.0, 28.0, 0.0),
       (1, 'Pane comune', 289, 8.6, 0.5, 63.5),
       (1, 'Latte intero', 64, 3.3, 3.6, 4.9);
SQL
```

Ora cerca "pasta", "mela", "parmigiano" → dovrebbe funzionare.

## Troubleshooting

**`go mod tidy` errori di import:**
- Spesso typo o package path errato. Mandami l'errore.

**`npm install` peer-dep warning:**
- React 19 + Tailwind 4 + Vite 6 sono recenti. `--legacy-peer-deps` se blocca.

**Migrate fail "Unknown collation uca1400_ai_ci":**
- MariaDB <11 non la supporta. Verifica `docker compose ps` mostri `mariadb:11.4`.

**Backend parte ma "connection refused":**
- DB non pronto. `docker compose logs mariadb` per vedere se è healthy.

**Camera/PWA non funzionano:**
- Servono HTTPS. In dev usa `mkcert` + Vite plugin per TLS locale, oppure testa solo le features non-camera.

## Reset completo

```bash
docker compose -f docker-compose.dev.yml down -v   # cancella DB
rm -rf bin/ web/node_modules web/dist .cache/      # cancella build + cache
```

## Stop normale

```bash
docker compose -f docker-compose.dev.yml down      # mantiene i dati
```
