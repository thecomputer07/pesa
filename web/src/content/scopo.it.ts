/**
 * Contenuto inline della pagina /scopo (IT).
 * TODO(phase-4): sostituire con import da `vite-plugin-md` per evitare drift.
 */
export const scopoIT = `# Chi siamo / Scopo

## Chi siamo

Pesa è un progetto open-source creato e mantenuto da Lorenzo Di Mare, sviluppatore indipendente. Il codice è pubblicato con licenza **AGPL v3** su repository pubblico: https://github.com/thecomputer07/pesa. Non c'è un'azienda, non c'è un investitore, non c'è pubblicità, non c'è un piano a pagamento: Pesa è fatto per chi conta i carboidrati e vuole uno strumento semplice, trasparente e costruito su dati pubblici ufficiali.

## A cosa serve Pesa

Pesa calcola i **carboidrati** di qualsiasi alimento in base al peso in grammi. Scrivi il nome dell'alimento, inserisci il peso, e ottieni il valore di carboidrati — più fibre, zuccheri e calorie come informazioni contestuali. I dati vengono da banche dati pubbliche ufficiali: prima di tutto la tabella italiana CREA; per ciò che CREA non copre, la tabella francese Ciqual e quella statunitense USDA, con i nomi tradotti in italiano.

Puoi usare Pesa per:

- contare i carboidrati di un pasto per motivi nutrizionali (sport, regime alimentare personale, curiosità);
- costruire un diario alimentare giornaliero e settimanale, offline-first;
- creare ricette composte da più ingredienti e riutilizzarle come "alimento composto" nel diario;
- scansionare il codice a barre di un prodotto confezionato al supermercato.

Pesa funziona **offline**: una volta caricato il catalogo alimenti sul tuo dispositivo, puoi cercare e calcolare anche senza connessione. Il diario si sincronizza quando torni online.

## A cosa NON serve Pesa

Pesa **non è un'app clinica**, non dà consigli sanitari, **non sostituisce il parere del medico, del nutrizionista o del dietista**. Mostra valori nutrizionali pubblici — non interpreta la tua situazione personale.

Se hai esigenze cliniche specifiche — ad esempio diabete, fenilchetonuria, celiachia diagnosticata, insufficienza renale, un regime alimentare prescritto da un professionista — è il tuo medico curante la persona giusta con cui discutere scelte alimentari e obiettivi. Pesa può starti a fianco come strumento di calcolo generico, nello stesso modo in cui una bilancia da cucina o una tabella nutrizionale cartacea stanno a fianco a chiunque: non prende decisioni al posto tuo, e chi ti segue clinicamente deve sapere cosa stai usando.

## Fonti dei dati

Pesa integra quattro fonti, tutte citate in modo visibile per ogni alimento:

- **CREA — Alimenti e Nutrizione** (alimentinutrizione.it): standard governativo italiano, circa 830 alimenti di base. Riuso consentito citando la fonte (CC BY 3.0).
- **Ciqual — ANSES** (ciqual.anses.fr): tabella francese di composizione degli alimenti, usata solo per ciò che CREA non copre — piatti composti (sushi, pizze, panini, zuppe), dolci, gelati e salse. Nomi tradotti in italiano da Pesa, valori invariati. Licenza Ouverte / Etalab 2.0.
- **USDA FoodData Central** (fdc.nal.usda.gov): banca dati del Dipartimento dell'Agricoltura degli Stati Uniti, usata solo per i tipi di sushi (maki, nigiri, uramaki) e la cucina giapponese. Nomi tradotti in italiano da Pesa; i carboidrati sono riportati come "disponibili" (fibre escluse) come in CREA. Pubblico dominio.
- **Open Food Facts** (openfoodfacts.org): banca dati collaborativa di prodotti confezionati con barcode, licenza ODbL. Usata per lo scanner.

Per ogni risultato di ricerca e per ogni scheda alimento, Pesa mostra un badge con la fonte del dato e un link alla pagina originale. C'è una pagina "Fonti" dedicata che riporta il credit esteso e tutte le licenze applicabili.

## Privacy in due righe

- Nessun tracker di terze parti, nessun Google Analytics, nessun pixel pubblicitario.
- Nessun dato sanitario: Pesa non chiede né memorizza informazioni sulla tua salute.
- Solo i dati necessari: email, password (sempre hashata), diario alimentare che hai scelto di salvare.
- Hosting in Italia (OVH Milano), backup cifrati in UE (Hetzner Finlandia); traffico protetto da Cloudflare.

Leggi l'[Informativa Privacy](pesadia.com/privacy), i [Termini di Servizio](pesadia.com/terms) e la [Cookie Policy](pesadia.com/cookies) per i dettagli. Puoi in qualunque momento esportare o cancellare il tuo account.

## Contatti

Per istanze GDPR (accesso, rettifica, cancellazione, portabilità): \`privacy@pesadia.com\`. Trovi la procedura completa nel documento \`subject-rights-procedure.md\`, che risponde entro **30 giorni**.

Per bug, proposte, contributi al codice: il repository pubblico https://github.com/thecomputer07/pesa con issue tracker. Per segnalazioni di sicurezza: \`pesadia.com/.well-known/security.txt\`.

Per questioni generali: lorenzo.dimare@icloud.com.

---

*Ultimo aggiornamento: 2026-04-13 — v1.0.*
`;
