# Cookie Policy — Pesa

**Versione:** 1.0
**Data ultimo aggiornamento:** 2026-04-13
**Lingua autoritativa:** italiano

Conformità di riferimento: **Garante per la protezione dei dati personali, Provvedimento n. 231 del 10 giugno 2021 — "Linee guida cookie e altri strumenti di tracciamento"**, oltre al Regolamento (UE) 2016/679 (GDPR) e alla Direttiva 2002/58/CE ("ePrivacy") come recepita in Italia dal D.Lgs. 196/2003 e D.Lgs. 101/2018.

## 1. Premessa

Un **cookie** è una stringa di testo, generalmente di piccole dimensioni, che un sito web invia al browser dell'utente e che il browser conserva per poi ritrasmetterla al sito stesso alle visite successive. Per estensione, il termine "cookie" comprende — ai fini della disciplina del Garante 2021 — anche gli **strumenti di tracciamento analoghi**: identificatori memorizzati in `localStorage` / `sessionStorage` / IndexedDB, fingerprint del browser/dispositivo, pixel tag, e tecnologie equivalenti utilizzate per identificare l'utente o il suo dispositivo.

Il servizio Pesa utilizza **esclusivamente cookie tecnici** strettamente necessari all'erogazione del servizio o alla sua sicurezza. In conformità al **Garante 2021 §4.1** e all'art. 122 comma 1 del D.Lgs. 196/2003, per i cookie tecnici **non è richiesto il consenso dell'interessato**: è sufficiente la presente **informativa**. Non è pertanto mostrato alcun banner di consenso (che sarebbe ingannevole e superfluo).

Pesa **non utilizza**: cookie di profilazione, cookie di marketing, cookie pubblicitari, cookie di social plugin, cookie di terze parti, pixel di tracciamento, fingerprinting, social login, CDN che memorizzino dati utente. **Zero tracker terzi. Zero banner.**

## 2. Elenco dei cookie utilizzati

| Nome | Tipo | Finalità | Durata | Attributi | Base giuridica |
|------|------|----------|--------|-----------|----------------|
| `pesa_session` | **Tecnico — sessione** | Mantenere la sessione autenticata dell'utente fra richieste HTTP. Contiene un identificativo **opaque** associato lato server a uno store `scs` su PostgreSQL; nessun dato personale viene memorizzato nel cookie stesso. | 30 giorni rolling (rinnovato a ogni richiesta autenticata); revoca immediata al logout | `HttpOnly`, `Secure`, `SameSite=Lax`, `Path=/` | **Art. 6(1)(b) GDPR** — esecuzione di un contratto; Garante 2021 §4.1 — cookie tecnico, no consenso |
| `pesa_csrf` | **Tecnico — sicurezza** | Protezione contro attacchi Cross-Site Request Forgery tramite pattern double-submit token (valore del cookie deve combaciare con header `X-CSRF-Token` inviato dal frontend nelle richieste mutanti) | Durata della sessione | `Secure`, `SameSite=Lax`, `Path=/`, **NON** `HttpOnly` (deve essere leggibile dal frontend) | **Art. 6(1)(f) GDPR** — legittimo interesse alla sicurezza |

**Nessun altro cookie è impostato dal servizio.** Nessun cookie persistente di profilazione, nessun cookie di terze parti, nessun tracker pubblicitario, nessun identificatore di prima o terza parte per fini di marketing.

### Storage locale aggiuntivo (non-cookie)

Per finalità strettamente tecniche, l'app può utilizzare:

- **IndexedDB / Dexie** per cachare il **dataset alimenti** in locale (offline-first). Non contiene identificatori personali; è svuotabile dall'utente via impostazioni del browser o via funzione "pulisci cache" dell'app quando disponibile.
- **Service Worker cache** (Workbox 7) per la shell PWA e gli asset statici. Nessun dato personale.

Questi storage non costituiscono "cookie" nel senso stretto dell'art. 122 del Codice Privacy, ma — in conformità al Garante 2021 §2 ("altri strumenti di tracciamento") — sono comunque dichiarati per trasparenza. Essendo **strettamente necessari all'erogazione** del servizio offline-first, ricadono nella stessa esenzione da consenso dei cookie tecnici.

## 3. Analytics self-hosted (Plausible Community Edition)

Qualora attivato (INFRA-14), il servizio utilizza **Plausible Community Edition**, installato sulla stessa infrastruttura del Titolare (container Docker accanto al binario Pesa e al database PostgreSQL). Plausible CE è uno strumento di analytics di prima parte con le seguenti caratteristiche:

- **Non utilizza cookie** — nessun identificativo viene memorizzato nel browser dell'utente;
- **Non utilizza localStorage / sessionStorage / IndexedDB** per analytics;
- **Non utilizza fingerprinting** del dispositivo o del browser;
- **Non memorizza l'indirizzo IP** in chiaro (utilizza un hash giornaliero ruotato per contare unique visitors entro la finestra giornaliera, senza persistere l'identificatore);
- **Non traccia l'utente tra siti** (first-party only);
- **Aggrega** unicamente: conteggio pageview, referrer generico, tipo di device (desktop / mobile / tablet), famiglia di browser, paese a livello aggregato.

**Nessun identificatore utente.** **Nessuna profilazione.** **Nessun collegamento fra visite.**

**Base giuridica:** Art. 6(1)(f) GDPR — legittimo interesse del Titolare alla conoscenza aggregata dell'utilizzo del servizio. Conformità al **Considerando 26 GDPR** (i dati completamente anonimi non rientrano nel perimetro GDPR) e al **Garante 2021 §4.2** (analytics anonimizzati di prima parte che non rendano il visitatore identificabile sono equiparati ai cookie tecnici e **non richiedono consenso**).

L'utente può comunque **opporsi** al conteggio futuro delle proprie visite scrivendo a `privacy@pesadia.com`: dal momento della richiesta, Plausible verrà configurato lato server per escludere le interazioni provenienti dall'IP/ambiente indicato (nota tecnica: poiché l'aggregato storico non contiene identificatori, non è possibile rimuovere retroattivamente specifiche interazioni dall'aggregato perché già anonime e non isolabili).

Dashboard di Plausible: accesso riservato esclusivamente al Titolare tramite autenticazione; nessun terzo ha accesso.

## 4. Come disabilitare i cookie

L'utente può in ogni momento disabilitare o cancellare i cookie intervenendo sulle impostazioni del proprio browser. Link alle guide ufficiali:

- **Google Chrome:** https://support.google.com/chrome/answer/95647
- **Mozilla Firefox:** https://support.mozilla.org/it/kb/Gestione%20dei%20cookie
- **Safari (macOS):** https://support.apple.com/it-it/guide/safari/sfri11471/mac
- **Safari (iOS):** https://support.apple.com/it-it/HT201265
- **Microsoft Edge:** https://support.microsoft.com/it-it/microsoft-edge/eliminare-i-cookie-in-microsoft-edge-63947406-40ac-c3b8-57b9-2a946a29ae09

**Avvertenza:** disabilitando i cookie tecnici `pesa_session` e `pesa_csrf`, l'autenticazione non funziona e le richieste mutanti vengono bloccate dalla protezione CSRF; il servizio resta comunque utilizzabile in **modalità demo** (ricerca alimenti + calcolo carboidrati senza account e senza diario sincronizzato), che non richiede sessione autenticata.

## 5. Modifiche

Ogni variazione sostanziale dell'elenco dei cookie (aggiunta di un nuovo cookie, modifica di finalità o durata) sarà pubblicata in questa pagina con **bump di versione** e, ove coinvolga cookie non-tecnici, con notifica in-app e richiesta di consenso all'utente. **L'introduzione di qualsiasi cookie di profilazione** richiederebbe banner di consenso ex Garante 2021 §2-§3 — una simile introduzione **non è prevista** nell'MVP e costituirebbe una decisione esplicita del Titolare documentata nel `PROJECT.md` Key Decisions.

Cronologia versioni:

- **v1.0** — 2026-04-13 — prima pubblicazione (Phase 0), esclusivamente cookie tecnici `pesa_session` e `pesa_csrf`.

## 6. Contatti

- **Titolare:** Lorenzo Di Mare
- **Email istanze privacy:** `privacy@pesadia.com`
- **Email generale:** lorenzo.dimare@icloud.com
- **PEC:** —
- **Autorità di controllo:** Garante per la protezione dei dati personali — Piazza Venezia 11, 00187 Roma — `garante@gpdp.it` — `protocollo@pec.gpdp.it` — https://www.garanteprivacy.it
