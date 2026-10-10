# Privacy Policy — Pesa

**Version:** 1.0
**Last updated:** 2026-04-13
**Authoritative language:** Italian

> *This is a convenience translation. In case of conflict with the Italian version (`privacy-policy.it.md`), the Italian version prevails.*

This notice is provided pursuant to Articles 13 and 14 of Regulation (EU) 2016/679 (GDPR) to users of Pesa — a generic nutritional carb calculator. It applies to the domain pesadia.com, the Pesa Progressive Web App, and the versioned HTTP API `/api/v1/` served from the same domain.

## 1. Data controller

- **Controller:** Lorenzo Di Mare
- **Email:** lorenzo.dimare@icloud.com
- **Certified email (PEC):** —
- **Data subject rights channel:** `privacy@pesadia.com`
- **DPO:** not appointed — reasoned justification at §11

The controller is a single natural person established in Italy; there are no joint controllers (Art. 26 GDPR); no EU representative is required (Art. 27 GDPR).

## 2. Nature of the service

Pesa is a **generic nutritional carb calculator** that computes carbohydrates (and, contextually, fiber/sugars/kcal) of foods contained in certified Italian databases (CREA) and in Open Food Facts. Extended positioning is documented in `.planning/legal/scope.md` (user-facing version: "About / Scope" page).

**Pesa is NOT a medical tool, is NOT a medical device under Regulation (EU) 2017/745 (MDR), does not provide diagnosis, therapy, or clinical advice.** Consistently, the service **does NOT process Art. 9 GDPR data** (health, biometric, genetic, sexual life, racial/ethnic origin, religious belief, trade-union membership, political opinion). None of these categories are collected, processed, inferred, or deduced.

## 3. Data processed

| Category | Fields | Origin | Retention |
|----------|--------|--------|-----------|
| **Account** | email, password hash (Argon2id), account creation date, preferred language, email-verified flag, (optional) date of birth not used in the MVP | provided by the user at signup | until account deletion |
| **Email verification & password reset tokens** | opaque pseudorandom token, expiry, associated user_id | server-generated | 24h (email verify), 1h (password reset); deleted after use or expiry |
| **Personal data export** (PRIV-01) | ZIP archive with JSON/CSV dump of the user's data, HMAC-signed download token | server-generated on user request | **7 days** from generation; download link auto-expires |
| **Sessions** | session ID (cookie `HttpOnly`, `Secure`, `SameSite=Lax`), User-Agent, pseudonymized IP (SHA-256 salted hash), last-access timestamp | server-generated on authentication | 30-day rolling; immediate revocation at logout |
| **Food diary** | meal entries (timestamp, food_id, weight in grams, computed CHO/fiber/sugars/kcal, meal type, optional free-text note), composed recipes (name, servings, ingredients), favorites | provided by the user | until account deletion or selective deletion (soft-delete + hard-delete after 30 days) |
| **Security logs** | pseudonymized IP, event type (login ok/failed, reset, rate-limit, 4xx/5xx, sensitive events: password/email change, 2FA, account delete, export), timestamp, req-ID | server-generated | **90-day target**, enforced via `logrotate` at deployment level |
| **Dedicated sensitive-events audit log (SEC-16)** | DB table storing auth/account event history | server-generated | **planned post-MVP v0.1.0 — not active in the MVP**; events are covered by the security logs above (90-day retention) |

**Important note:** no special category of data under **Art. 9 GDPR** (health, biometric, genetic, etc.) is collected, inferred, or deduced. The app **does not ask for, infer, or store** any information about the user's health. The food diary is user-generated content analogous to a personal diary; the service does not classify users by health condition, does not perform clinical pattern recognition, does not produce health inferences.

## 4. Legal bases

| Processing | GDPR legal basis | Reference |
|------------|------------------|-----------|
| Account management (email, password, signup, login, reset) | **Art. 6(1)(b)** — performance of a contract requested by the data subject | RoPA Activity A |
| Food diary, recipes, favorites, cross-device sync | **Art. 6(1)(b)** — performance of a contract requested by the data subject | RoPA Activity B |
| Session technical cookie (`pesa_session`) | **Art. 6(1)(b)** — strictly necessary; Italian Garante 2021 no. 231 §4.1 (no consent required) | Cookie Policy §2 |
| CSRF technical cookie (`pesa_csrf`) | **Art. 6(1)(f)** — legitimate interest in service security | Cookie Policy §2 |
| Security logs + audit log | **Art. 6(1)(f)** — legitimate interest in correct functioning and security, balanced with pseudonymization and proportional retention | RoPA Activity C |
| Plausible Community Edition self-hosted analytics (if active) | **Art. 6(1)(f)** — legitimate interest in aggregate service usage knowledge, no cookies, no fingerprinting, no user identifiers; GDPR Recital 26 and Garante 2021 §4.2 compliance (first-party anonymized analytics) | Cookie Policy §3 |

**No processing is based on Art. 9 GDPR.** No processing requires consent under Art. 6(1)(a) in the MVP.

## 5. Recipients

Personal data are not disclosed to third parties for their own purposes. The following infrastructure providers are designated as **processors under Art. 28 GDPR**:

- **OVH Italia S.r.l.** — primary hosting (VPS Milan, Italy) for application server and PostgreSQL database. Signed DPA archived in `.planning/legal/dpa/`.
- **Hetzner Online GmbH** — encrypted backups (servers in Finland and/or Germany). Signed DPA archived in `.planning/legal/dpa/`.

Both providers are established in the European Union; both operate exclusively on EU data centers; extra-EEA sub-transfers are not authorized for Pesa data.

**No other recipients.** No ad broker, no third-party analytics, no CDN carrying user data, no marketing platform, no social login, no pixel. Third-party data sources (CREA, Open Food Facts) are **consulted outbound** — Pesa fetches public food information from them without sending any user data.

Communications to public authorities: only upon formal motivated request, within the limits of law, with notice to the data subject where permitted.

## 6. Retention periods

The table in §3 provides category-level retention. In summary:

- Account and diary: **until deletion** is requested by the user (self-service PRIV-05 from Phase 6 onward, or email to `privacy@pesadia.com` per `.planning/legal/subject-rights-procedure.md`).
- Hetzner backups: encrypted snapshots with **30-day retention**; a **tombstone catalog** (PRIV-04) guarantees that deletion requests propagate to backups within 30 days.
- Application security logs: **90-day target** (proportionality, Art. 5(1)(c) GDPR), enforced via `logrotate` at deployment level.
- Dedicated sensitive-events audit log (SEC-16): **planned post-MVP v0.1.0**; in the MVP, sensitive events are tracked in the application slog logs with 90-day retention. When activated, target retention is **12 months** (Art. 17(3)(e) GDPR).
- Subject-rights ticket archive: **3 years** (compliance evidence for Garante), with hashed identifiers — see procedure §8.
- Email verification token: **24h**. Password reset token: **1h**.
- Personal data export (ZIP archive + signed link): **7 days** from generation.
- Anonymized erasure log (`erasure_log`): **12-month target retention**; in the MVP v0.1.0 rotation is manual, automation planned post-MVP.

## 7. Data subject rights

Within the limits of Articles **15-22 GDPR**, the data subject has the right to:

- **access** their data (Art. 15);
- obtain **rectification** of inaccurate data (Art. 16);
- obtain **erasure** ("right to be forgotten", Art. 17);
- obtain **restriction** of processing (Art. 18);
- receive their data in a **portable**, structured, commonly-used, machine-readable format (Art. 20);
- **object** to processing based on legitimate interest on grounds relating to their particular situation (Art. 21);
- not be subject to decisions based **solely on automated processing**, including profiling (Art. 22) — for Pesa this right is already satisfied de facto: the service performs **no** automated decision-making or profiling.

**Consent withdrawal:** not applicable in the MVP, as no processing is based on Art. 6(1)(a).

**Exercising rights:** the user writes to `privacy@pesadia.com`. The procedure is documented in `.planning/legal/subject-rights-procedure.md`. Ordinary SLA **30 days** from receipt; extendable by 60 additional days (total 90) with motivated notice within the first 30 days under Art. 12(3) GDPR.

**Complaint to the supervisory authority:** the data subject always has the right to lodge a complaint under Art. 77 GDPR:

- **Garante per la protezione dei dati personali** — Piazza Venezia 11, 00187 Rome, Italy
- Email: `garante@gpdp.it` — Certified: `protocollo@pec.gpdp.it`
- Website: https://www.garanteprivacy.it

## 8. Security

The controller adopts technical and organizational measures appropriate under Art. 32 GDPR:

- **Passwords** — Argon2id hashing with OWASP 2025 parameters (m=64 MiB, t=3, p=4), random per-password salt; constant-time comparison; NIST 800-63B 2024 policy (min 12 chars, no forced expiration, common-password block via offline HaveIBeenPwned).
- **Transport** — TLS 1.3 only (TLS 1.2 fallback for compatibility only), HSTS preload, Mozilla "modern" cipher suites, OCSP stapling.
- **Application security** — strict CSP with no `unsafe-inline`/`unsafe-eval`, restrictive Permissions-Policy, COOP/COEP/CORP, `X-Content-Type-Options: nosniff`.
- **Database** — internal TLS Go↔PostgreSQL even on localhost; at-rest data cluster encryption via LUKS; least-privilege DB users (`pesa_app`, `pesa_migration`, `pesa_backup`, `pesa_admin`).
- **Sessions** — opaque server-side session IDs (scs + PostgreSQL store); regenerated on privilege change; CSRF double-submit tokens.
- **Authentication** — account lockout after 5 failed attempts, rate limiting on signup/login/reset, email-enumeration prevention.
- **Backups** — AES-256 encrypted pg_basebackup with separated key, destination Hetzner (FI/DE), quarterly documented restore drill.
- **Audit** — slog security logs with 90-day target retention (pseudonymized IP) also cover sensitive events in the MVP v0.1.0; dedicated audit log (SEC-16, 12-month retention) planned post-MVP.
- **Development** — continuous dependency scanning (govulncheck, osv-scanner, npm audit), SAST (gosec, semgrep), secrets scanning (gitleaks), CycloneDX SBOM per release, GPG-signed releases, RFC 9116 `security.txt`.
- **Operations** — SSH key-based only, restrictive firewall, unattended-upgrades, monitoring with alerting on 5xx spikes, failed-login spikes, disk usage, backup failures.

## 9. Extra-EEA transfers

**No personal data transfer occurs outside the European Economic Area.** Hosting OVH Milan (Italy), backups Hetzner (Finland/Germany) — both EU. The consulted third-party sources (CREA — Italian public body; Open Food Facts — France) are also EU-based; for OFF, the barcode call uses a custom User-Agent with no user identifier.

Should an extra-EEA transfer become necessary in the future, the safeguards under Arts. 45-49 GDPR will apply (adequacy decision or standard contractual clauses with post-Schrems II complementary assessment), and this policy will be updated with at least 30 days' notice.

## 10. Minors

The service is intended for a general audience aged **14 and above**, consistent with the threshold set by **Italian Legislative Decree 101/2018 Art. 2-quinquies** implementing Art. 8 GDPR in Italy. At signup the user declares to be at least 14 via informed consent (AUTH-07) — non-pre-ticked checkbox.

Since no direct processing of data of minors under 14 is envisaged, the reinforced protection under Art. 8 GDPR (parental authorization) is not necessary. No feature of the app is designed for, or oriented to, minors under 14.

## 11. DPO — justification for non-appointment

The controller has **not appointed** a Data Protection Officer (DPO) under Art. 37 GDPR. Non-appointment is justified as follows, addressing each sub-paragraph of Art. 37(1):

- **(a)** processing is carried out by a **private party** (single natural person), not by a public authority or body → point (a) not applicable;
- **(b)** the core activity does **not consist** of processing operations that, by their nature, require **regular and systematic monitoring of data subjects on a large scale** — no tracking, no profiling, no pattern recognition; planned MVP user base below 10,000 → point (b) not met;
- **(c)** the core activity does **not consist** of large-scale processing of special categories (Art. 9) or criminal conviction data (Art. 10) — **no Art. 9 data** processed by design (see §3) → point (c) not met.

None of the three conditions is met: appointment is not mandatory. The controller remains the single contact point for all privacy requests at `privacy@pesadia.com` and PEC `—`. **Voluntary** appointment will be reassessed if the user base exceeds 10,000 subjects or the characteristics of processing change.

## 12. Changes to this notice

Any substantive change to the notice (new purpose, new legal basis, new recipient, new transfer, retention change, introduction of required consent) will entail:

- version bump (e.g. v1.0 → v1.1 for clarifications, v2.0 for structural changes);
- update of the "last updated" date;
- in-app notification to registered users **at least 15 days before** the effective date;
- courtesy email to users with an active account.

Minor editorial changes (typos, linguistic clarifications without substantive effect) are published without prior notice but tracked in git.

## 13. Contacts

- **Controller:** Lorenzo Di Mare
- **GDPR requests email:** `privacy@pesadia.com`
- **General email:** lorenzo.dimare@icloud.com
- **Certified email (PEC):** —
- **Complaint:** Garante per la protezione dei dati personali — Piazza Venezia 11, 00187 Rome, Italy — `garante@gpdp.it` — `protocollo@pec.gpdp.it` — https://www.garanteprivacy.it

## 14. Versioning

- **v1.0** — 2026-04-13 — first publication (Phase 0).

Full diff tracked in the public repository https://github.com/thecomputer07/pesa under `.planning/legal/privacy-policy.en.md`.
