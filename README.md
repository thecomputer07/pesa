# Pesa

Calcolatore nutrizionale generico (carboidrati / CHO) basato su fonti alimentari italiane (CREA + BDA IEO).

> **Disclaimer importante:** Pesa NON è un dispositivo medico, NON sostituisce il parere di un professionista sanitario, NON tratta dati di categoria particolare (GDPR Art. 9). Vedi `.planning/legal/scope.md` per la dichiarazione completa.

---

## Scope

Inserisci un alimento + peso in grammi → ottieni i grammi di carboidrati in <5s, anche offline (PWA). Italiano-first, EN come traduzione.

## Stack

- **Backend:** Go 1.22+, chi router, scs sessions su MariaDB, Argon2id
- **DB:** MariaDB 11.x con TLS, migrazioni Goose embedded
- **Frontend:** Vite + React + TypeScript (PWA in fasi successive)
- **Deploy:** binario singolo + systemd su OVH Milano, autocert TLS, backup cifrato GPG → Hetzner
- **Licenza:** AGPL-3.0-only

## Quick start (dev)

```bash
make dev          # air hot reload (richiede `go install github.com/cosmtrek/air@latest`)
make test         # unit + integration
make lint         # golangci-lint + gosec + govulncheck
cd web && npm install && npm run dev
```

## Struttura

| Dir | Scopo |
|---|---|
| `cmd/pesa/` | Entry point binario |
| `internal/{config,domain,repo,service,http,i18n,legal}/` | Codice applicativo |
| `migrations/` | Goose `.sql` |
| `web/` | Vite + React + TS skeleton |
| `deploy/` | systemd + docker-compose + backup |
| `docs/deploy/` | Runbook produzione |
| `.planning/legal/` | Documenti legali (privacy, ToS, RoPA, CREA outreach) |
| `MARKETING-RULES.md` | Lista parole proibite (CI scanner) |
| `scripts/` | render-placeholders, anti-word-scan, reproducible-build |

## Contribuire

Vedi `CONTRIBUTING.md`. DCO sign-off richiesto, no CLA. Codice sotto AGPL-3.0.

## Sicurezza

Vulnerabilità → `SECURITY.md` (GPG-encrypted email, SLA 7gg).

## Status

Pre-MVP. Vedi `.planning/ROADMAP.md` per milestone v0.1.0.
