# Informativa Privacy — Pesa

**Versione:** 1.1
**Data ultimo aggiornamento:** 2026-10-12
**Lingua autoritativa:** italiano

La presente informativa è resa ai sensi degli artt. 13 e 14 del Regolamento (UE) 2016/679 (GDPR) agli utenti del servizio Pesa — calcolatore nutrizionale generico. Si applica al dominio pesadia.com, all'applicazione web (PWA) Pesa e all'API HTTP versionata `/api/v1/` erogate dallo stesso dominio.

## 1. Titolare del trattamento

- **Titolare:** Lorenzo Di Mare
- **Email:** lorenzo.dimare@icloud.com
- **PEC:** —
- **Canale istanze interessato:** `privacy@pesadia.com`
- **DPO:** non designato — giustificazione motivata alla §11

Il titolare è persona fisica singola stabilita in Italia; non sono presenti contitolari ex art. 26 GDPR; non è richiesto rappresentante UE ex art. 27 GDPR.

## 2. Natura del servizio

Pesa è un **calcolatore nutrizionale generico** che calcola i carboidrati (e, come contorno, fibre/zuccheri/kcal) di alimenti presenti in banche dati italiane certificate (CREA) e in Open Food Facts. Per il posizionamento esteso si rimanda alla pagina "Chi siamo / Scopo" del servizio.

**Pesa NON è uno strumento medico, non è un dispositivo medico ai sensi del Regolamento UE 2017/745, non fornisce diagnosi, terapie, consigli clinici.** Coerentemente, il servizio **NON tratta dati di cui all'art. 9 GDPR** (dati sanitari, biometrici, genetici, relativi alla vita sessuale, all'origine etnica, alle convinzioni religiose, all'appartenenza sindacale, alle opinioni politiche). Nessuna di queste categorie viene raccolta, elaborata, inferita o dedotta.

## 3. Dati trattati

| Categoria | Campi | Origine | Conservazione |
|-----------|-------|---------|---------------|
| **Account** | email, password hash (Argon2id), data creazione account, lingua preferita, flag email verificata, (facoltativa) data di nascita non utilizzata nell'MVP | conferiti dall'utente al signup | fino alla cancellazione dell'account |
| **Verifica email e reset password** | token pseudocasuale opaque, scadenza, user_id associato | generati server-side | 24h (verifica email), 1h (reset password); eliminati dopo uso o scadenza |
| **Export dati personali** (PRIV-01) | archivio ZIP con dump JSON/CSV dei dati dell'utente, token di download firmato HMAC | generati server-side su richiesta dell'utente | **7 giorni** dalla generazione; link download scade automaticamente |
| **Sessioni** | session ID (cookie `HttpOnly`, `Secure`, `SameSite=Lax`), User-Agent, IP pseudonimizzato (hash SHA-256 salted), timestamp ultimo accesso | generati server-side all'autenticazione | 30gg rolling; revoca immediata al logout |
| **Diario alimentare** | entries pasto (timestamp, food_id, peso in grammi, CHO/fibre/zuccheri/kcal calcolati, tipo pasto, nota testuale facoltativa dell'utente), ricette composte (nome, porzioni, ingredienti), preferiti | conferiti dall'utente | fino a cancellazione dell'account o cancellazione selettiva (soft-delete + hard-delete dopo 30gg) |
| **Log di sicurezza** | IP pseudonimizzato, tipo evento (login ok/fallito, reset, rate-limit, 4xx/5xx, eventi sensibili: password/email change, 2FA, account delete, export), timestamp, req-ID | server-side | **90 giorni**, rotazione automatica sul server |
| **Audit log eventi sensibili dedicato (SEC-16)** | tabella DB con storia eventi auth/account | server-side | **pianificato post-MVP v0.1.0 — nell'MVP non attivo**; gli eventi sono coperti dai log di sicurezza (riga sopra) con retention 90gg |

**Nota importante:** nessuna categoria particolare ex **art. 9 GDPR** (dati sanitari, biometrici, genetici, ecc.) viene raccolta, inferita o dedotta. L'app **non chiede, non deduce, non memorizza** alcuna informazione sulla salute dell'utente. Il diario alimentare è contenuto generato dall'utente, equiparabile a un diario personale; il servizio non classifica utenti in base a condizioni di salute, non esegue pattern recognition clinico, non produce inferenze sanitarie.

## 4. Basi giuridiche

| Trattamento | Base giuridica GDPR | Riferimento |
|-------------|--------------------|-----------| 
| Gestione account (email, password, signup, login, reset) | **Art. 6(1)(b)** — esecuzione di un contratto di servizio richiesto dall'interessato | RoPA Attività A |
| Diario alimentare, ricette, preferiti, sincronizzazione | **Art. 6(1)(b)** — esecuzione di un contratto di servizio richiesto dall'interessato | RoPA Attività B |
| Cookie tecnici di sessione (`pesa_session`) | **Art. 6(1)(b)** — strettamente necessari all'erogazione del servizio; Garante 2021 n. 231 §4.1 (no consenso richiesto) | Cookie Policy §2 |
| Cookie tecnico CSRF (`pesa_csrf`) | **Art. 6(1)(f)** — legittimo interesse alla sicurezza del servizio | Cookie Policy §2 |
| Log di sicurezza + audit log | **Art. 6(1)(f)** — legittimo interesse al corretto funzionamento e sicurezza del servizio, bilanciato con pseudonimizzazione e retention proporzionata | RoPA Attività C |
| Analytics Plausible Community Edition self-hosted (se attivi) | **Art. 6(1)(f)** — legittimo interesse alla conoscenza aggregata dell'utilizzo del servizio, senza cookie, senza fingerprint, senza identificatori utente; conforme Considerando 26 GDPR e Garante 2021 §4.2 (analytics anonimizzati di prima parte) | Cookie Policy §3 |

**Nessun trattamento è fondato sull'art. 9 GDPR.** Nessun trattamento richiede consenso ex art. 6(1)(a) nell'MVP.

## 5. Destinatari

I dati personali non sono comunicati a terzi per finalità proprie. Sono designati **responsabili del trattamento ex art. 28 GDPR** i seguenti fornitori di infrastruttura:

- **OVH Italia S.r.l.** — hosting primario (VPS Milano, Italia) per server applicativo e database PostgreSQL.
- **Hetzner Online GmbH** — server dedicato del titolare a Helsinki (Finlandia) che riceve le copie di backup cifrate e invia le email di servizio (verifica indirizzo, reset password).
- **Cloudflare, Inc.** — DNS e reverse proxy/CDN davanti a pesadia.com: tutto il traffico HTTPS transita dalla rete Cloudflare, che vede indirizzo IP, User-Agent e contenuto delle richieste per instradarle e proteggerle da abusi; mette in cache solo contenuti pubblici uguali per tutti (file dell'app, foto e catalogo alimenti), mai dati dell'account o del diario.

Con OVH e Cloudflare l'accordo sul trattamento dei dati ex art. 28 GDPR (DPA) è incluso nelle rispettive condizioni contrattuali. Con Hetzner il rapporto è regolato dalle condizioni generali del servizio di server dedicato; i dati che vi transitano sono backup cifrati (chiave non conservata presso Hetzner) ed email di servizio. OVH e Hetzner trattano i dati Pesa esclusivamente in data center UE; per Cloudflare vedi §9.

**Nessun destinatario ulteriore.** Nessun broker pubblicitario, nessun fornitore di analytics terzo, nessuna piattaforma di marketing, nessun social login, nessun pixel. Le fonti dati terze (CREA, Open Food Facts) sono **consultate in uscita** — Pesa richiede loro informazioni pubbliche su alimenti, senza inviare dati utente.

Comunicazioni a autorità pubbliche: solo su richiesta formale motivata, nei limiti di legge, con notifica all'interessato quando consentito.

## 6. Tempi di conservazione

La tabella alla §3 indica i tempi per singola categoria. In sintesi:

- Account e diario: **fino a cancellazione** richiesta dall'utente (self-service PRIV-05 dalla Phase 6, oppure email a `privacy@pesadia.com` secondo la procedura interna del titolare).
- Backup: copia giornaliera sul server Milano (**14 giorni**) e copia cifrata in Finlandia (**30 giorni**); un **tombstone catalog** (PRIV-04) garantisce che la cancellazione richiesta dall'utente si propaghi ai backup entro 30 giorni.
- Log applicativi e di accesso (metodo, percorso, esito, tempo di risposta; IP pseudonimizzato o assente): **90 giorni** (proporzionalità ex art. 5(1)(c) GDPR), con rotazione automatica sul server.
- Audit log eventi sensibili dedicato (SEC-16): **pianificato post-MVP v0.1.0**; nell'MVP gli eventi sono tracciati nei log applicativi slog con retention 90gg. Quando attivato, retention target **12 mesi** (Art. 17(3)(e) GDPR).
- Archivio ticket istanze interessato: **3 anni** (evidenza compliance Garante), con hash dei dati identificativi — vedi procedura §8.
- Token di verifica email: **24h**. Token di reset password: **1h**.
- Export dati personali (archivio ZIP + link firmato): **7 giorni** dalla generazione.
- Log anonimizzato erasure (`erasure_log`): **retention target 12 mesi**; nell'MVP v0.1.0 la rotazione è manuale, automazione pianificata post-MVP.

## 7. Diritti dell'interessato

L'interessato ha diritto, nei limiti e con le modalità previste dagli artt. **15-22 GDPR**, di:

- **accedere** ai dati che lo riguardano (art. 15);
- ottenere la **rettifica** dei dati inesatti (art. 16);
- ottenere la **cancellazione** dei dati ("diritto all'oblio", art. 17);
- ottenere la **limitazione** del trattamento (art. 18);
- ricevere i propri dati in formato **portabile** strutturato, di uso comune, leggibile da dispositivo automatico (art. 20);
- **opporsi** al trattamento fondato su legittimo interesse per motivi connessi alla sua situazione particolare (art. 21);
- non essere sottoposto a decisioni basate **unicamente su trattamenti automatizzati**, inclusa la profilazione (art. 22) — per Pesa questa tutela è di fatto già rispettata poiché il servizio **non effettua** alcun processo decisionale automatizzato né profilazione.

**Revoca del consenso:** non applicabile nell'MVP, poiché nessun trattamento è fondato su art. 6(1)(a).

**Esercizio dei diritti:** l'utente scrive a `privacy@pesadia.com`. SLA ordinario **30 giorni** dalla ricezione; prorogabile a 60 giorni aggiuntivi (totale 90) con notifica motivata entro i primi 30 giorni ex art. 12(3) GDPR.

**Reclamo al Garante:** l'interessato ha diritto di proporre reclamo all'Autorità di controllo ex art. 77 GDPR:

- **Garante per la protezione dei dati personali** — Piazza Venezia 11, 00187 Roma
- Email: `garante@gpdp.it` — PEC: `protocollo@pec.gpdp.it`
- Sito: https://www.garanteprivacy.it

## 8. Sicurezza

Il titolare adotta misure tecniche e organizzative adeguate ex art. 32 GDPR:

- **Password** — hashing Argon2id con parametri OWASP 2025 (m=64 MiB, t=3, p=4), salt random per password; confronto timing-safe; policy NIST 800-63B 2024 (minimo 12 caratteri, nessuna scadenza forzata, blocco password comuni tramite HaveIBeenPwned offline).
- **Trasporto** — HTTPS obbligatorio (TLS 1.2/1.3) con HSTS; certificati Let's Encrypt sul server e Cloudflare in ingresso.
- **Application security** — CSP strict senza `unsafe-inline`/`unsafe-eval`, Permissions-Policy restrittiva, COOP/COEP/CORP, `X-Content-Type-Options: nosniff`.
- **Database** — PostgreSQL raggiungibile solo dalla rete interna dei container, mai esposto su internet; connessioni applicazione↔database cifrate con TLS (le connessioni in chiaro sono rifiutate).
- **Sessioni** — session ID opaque server-side (scs + PostgreSQL store); regenerate su privilege change; CSRF double-submit token.
- **Autenticazione** — account lockout 5 tentativi falliti, rate limiting su signup/login/reset, prevenzione email enumeration.
- **Backup** — dump giornaliero verificato; la copia fuori sede (Finlandia) è cifrata AES-256 con chiave conservata separatamente; ripristino provato.
- **Audit** — log sicurezza slog con retention target 90gg (IP pseudonimizzato) coprono anche gli eventi sensibili nell'MVP v0.1.0; audit log dedicato (SEC-16, retention 12 mesi) pianificato post-MVP.
- **Sviluppo** — dependency scanning continuo (govulncheck, osv-scanner, npm audit), SAST (gosec, semgrep), secrets scanning (gitleaks), SBOM CycloneDX per release, release firmate GPG, `security.txt` ex RFC 9116.
- **Operativa** — SSH key-based only, firewall restrittivo, unattended-upgrades, monitoring con alerting su 5xx spike, login failed spike, disk usage, backup failure.

## 9. Trasferimenti extra-SEE

Hosting (OVH Milano, Italia) e backup/email (Hetzner, Finlandia) sono nell'UE. **Cloudflare, Inc.** ha sede negli Stati Uniti e gestisce una rete globale: il traffico può essere elaborato in data center fuori dallo SEE. Il trasferimento è fondato sulla decisione di adeguatezza UE-USA (EU-U.S. Data Privacy Framework, a cui Cloudflare aderisce) e, in via complementare, sulle clausole contrattuali tipo incluse nel DPA di Cloudflare (art. 46 GDPR). Le fonti dati terze consultate (CREA — ente pubblico italiano; Open Food Facts — Francia) sono anch'esse nell'UE; per OFF la chiamata barcode utilizza un User-Agent custom senza alcun identificatore utente.

Qualora in futuro divenisse necessario un trasferimento extra-SEE, si applicheranno le garanzie ex artt. 45-49 GDPR (decisione di adeguatezza o clausole contrattuali tipo con valutazione complementare post-Schrems II) e questa informativa sarà aggiornata con preavviso minimo 30 giorni.

## 10. Minori

Il servizio è rivolto a un pubblico generico di età ≥ **14 anni**, coerentemente con la soglia stabilita dal **D.Lgs. 101/2018 art. 2-quinquies** come recepita dell'art. 8 GDPR in Italia. Al signup l'utente dichiara di avere almeno 14 anni tramite consenso informato (AUTH-07) — checkbox non pre-selezionato.

Non essendo previsto alcun trattamento diretto di dati di minori sotto i 14 anni, non è necessaria la tutela rafforzata art. 8 GDPR (autorizzazione del titolare della responsabilità genitoriale). Nessuna funzione dell'app è pensata per, né orientata a, minori sotto i 14 anni.

## 11. DPO — giustificazione di non designazione

Il titolare **non ha designato** un Responsabile della Protezione dei Dati (DPO) ex art. 37 GDPR. La non-designazione è motivata come segue, punto per punto dell'art. 37(1):

- **(a)** il trattamento è svolto da un'**entità privata** (persona fisica singola), non da un'autorità pubblica o organismo pubblico → lettera (a) non applicabile;
- **(b)** l'attività principale **non consiste** in trattamenti che per loro natura richiedono il **monitoraggio regolare e sistematico su larga scala** degli interessati — non esiste tracking, non esiste profiling, non esiste pattern recognition; la base utenti prevista nell'MVP è inferiore a 10.000 soggetti → lettera (b) non soddisfatta;
- **(c)** l'attività principale **non consiste** in trattamenti **su larga scala** di categorie particolari ex art. 9 o di dati personali relativi a condanne penali ex art. 10 — **nessun dato art. 9** viene trattato per design (vedi §3) → lettera (c) non soddisfatta.

Nessuno dei tre presupposti è soddisfatto: la designazione non è obbligatoria. Il titolare resta punto di contatto unico per tutte le istanze privacy via `privacy@pesadia.com` e PEC `—`. Designazione **volontaria** sarà rivalutata quando la base utenti superi 10.000 soggetti o quando cambino le caratteristiche del trattamento.

## 12. Modifiche alla presente informativa

Ogni modifica sostanziale dell'informativa (nuova finalità, nuova base giuridica, nuovo destinatario, nuovo trasferimento, variazione della retention, introduzione di consenso richiesto) comporterà:

- bump di versione (es. v1.0 → v1.1 per chiarimenti, v2.0 per modifiche strutturali);
- aggiornamento della data ultimo aggiornamento;
- notifica in-app agli utenti registrati **almeno 15 giorni prima** dell'entrata in vigore;
- email di cortesia agli utenti con account attivo.

Modifiche editoriali minori (refusi, chiarimenti linguistici senza effetti sostanziali) vengono pubblicate senza preavviso ma tracciate in git.

## 13. Contatti

- **Titolare:** Lorenzo Di Mare
- **Email istanze GDPR:** `privacy@pesadia.com`
- **Email generale:** lorenzo.dimare@icloud.com
- **PEC:** —
- **Reclamo:** Garante per la protezione dei dati personali — Piazza Venezia 11, 00187 Roma — `garante@gpdp.it` — `protocollo@pec.gpdp.it` — https://www.garanteprivacy.it

## 14. Versioning

- **v1.0** — 2026-04-13 — prima pubblicazione (Phase 0).

Diff tracciato integralmente nel repository pubblico https://github.com/thecomputer07/pesa sotto `web/public/legal/privacy-it.md`.
