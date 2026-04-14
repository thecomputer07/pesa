# Contribuire a Pesa

Grazie per l'interesse. Pesa è AGPL-3.0 e accoglie contributi su:
- Bug fix
- Documentazione (IT/EN)
- Estensioni dataset alimentare (CREA, BDA IEO, OFF)
- Miglioramenti accessibilità (WCAG 2.2 AA)

## Workflow

1. Apri una issue prima per cambiamenti non banali — evita lavoro sprecato
2. Fork → branch `feat/<short-name>` o `fix/<short-name>` dalla `main`
3. Commit firmati DCO: `git commit -s -m "..."` (Developer Certificate of Origin, no CLA)
4. PR contro `main`. CI deve passare (vedi sotto).

## Standard codice

**Go:**
- `gofmt` + `goimports` (Makefile target `make fmt`)
- `golangci-lint run` deve passare (config in `.golangci.yml`)
- `gosec ./...` zero high/critical
- `govulncheck ./...` zero vulnerabilità note
- Test obbligatori per logica nuova (`go test ./...` deve passare)

**Frontend (web/):**
- TypeScript strict (`tsconfig.json`)
- ESLint + Prettier (config da definire in `web/`)
- Vitest unit + Playwright smoke

**Marketing rules (CRITICO):**
- `bash scripts/anti-word-scan.sh` deve passare prima della PR
- Nessuna parola medica/diagnostica/terapeutica nelle UI/copy/marketing
- Vedi `MARKETING-RULES.md` per la lista completa

## Commit message

Conventional Commits-style consigliato:
```
feat(auth): add Argon2id rehash on login
fix(http): correct Cache-Control on /healthz
docs: update deploy runbook
```

## DCO Sign-off

Aggiungi il sign-off con `-s`. Significa che attesti i criteri del Developer Certificate of Origin v1.1 (https://developercertificate.org/).

## Review

PR vengono revisionate da `Lorenzo Di Mare` o maintainer designati. Tempo target: 7 giorni feriali. Per PR di sicurezza vedi `SECURITY.md` (NON aprire issue pubblica).

## Licenza

Inviando una PR accetti che il tuo contributo sia rilasciato sotto AGPL-3.0-only, stessa licenza del progetto.
