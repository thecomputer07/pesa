# Security Policy

## Versioni supportate

| Versione | Supportata |
|---|---|
| `main` (HEAD) | ✓ |
| Ultima release stabile | ✓ |
| Release precedenti | ✗ |

## Reporting

**NON aprire issue pubbliche per vulnerabilità di sicurezza.**

Invia una mail cifrata a:

- **Email:** `security@pesadia.com` (placeholder fino a deploy)
- **GPG key fingerprint:** `{{GPG_FINGERPRINT}}` — chiave pubblica in `.well-known/pgp-key.txt`
- **Disclosure window:** rispondiamo entro **7 giorni feriali**, fix entro **30 giorni** per HIGH/CRITICAL

Includi:
1. Descrizione tecnica della vulnerabilità
2. Steps to reproduce
3. Impatto stimato (CVSS 3.1 se possibile)
4. PoC (opzionale)
5. Tuo contatto per follow-up + se vuoi credit pubblico

## Coordinated disclosure

Pratica preferita: 90 giorni embargo dalla data di reporting al disclosure pubblico, salvo accordi diversi.

## In scope

- Codice in questo repository (Go backend + web frontend)
- Configurazioni di deploy (`deploy/`)
- Migrazioni MariaDB
- CI workflows

## Out of scope

- Infrastruttura terza (OVH, Hetzner, MariaDB upstream, dipendenze npm/Go upstream — segnala upstream)
- Social engineering del maintainer
- DoS volumetrico
- Vulnerabilità che richiedono accesso fisico al server

## Acknowledgements

Reporter responsabili saranno menzionati in `SECURITY-CREDITS.md` (se danno consenso).

## security.txt

Vedi `.well-known/security.txt` per il formato RFC 9116.
