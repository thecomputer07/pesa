# Pesa — Template autovalutazione accessibilità (Legge Stanca / EAA)

Usa questo modello a ogni revisione periodica (minimo annuale).
Archivia il risultato compilato in `docs/a11y-audits/YYYY-MM-DD.md`.

## Metadati

- **Data audit:** YYYY-MM-DD
- **Versione applicativa:** {git tag o SHA}
- **Auditor:** {nome, ruolo}
- **Metodo:** autovalutazione (art. 3 Dir. UE 2016/2102)
- **Pagine campione:** `/`, `/cerca`, `/alimento/:id`, `/scopo`, `/accessibilita`, `/chi-siamo`

## Strumenti usati

| Strumento              | Versione | Scopo                                     |
| ---------------------- | -------- | ----------------------------------------- |
| axe-core               | 4.x      | Scansione automatica WCAG 2.2 AA          |
| NVDA                   | 2024.x   | Screen reader Windows                     |
| VoiceOver              | macOS 14 | Screen reader macOS/iOS                   |
| TalkBack               | Android  | Screen reader mobile                      |
| Firefox DevTools       | 128      | Simulatore contrasto + emulazione motion  |
| Tastiera fisica        | —        | Navigazione Tab/Shift+Tab/Enter/Esc       |

## Checklist WCAG 2.2 AA

Per ogni criterio, annota: PASS / FAIL / N/A + nota.

| Criterio | Descrizione                                | Esito | Nota |
| -------- | ------------------------------------------ | ----- | ---- |
| 1.1.1    | Contenuto non testuale                     |       |      |
| 1.3.1    | Info e relazioni                           |       |      |
| 1.4.3    | Contrasto (minimo)                         |       |      |
| 1.4.4    | Ridimensionamento testo                    |       |      |
| 1.4.11   | Contrasto non testuale                     |       |      |
| 2.1.1    | Tastiera                                   |       |      |
| 2.1.2    | Nessun trap da tastiera                    |       |      |
| 2.4.1    | Bypass blocks (skip link)                  |       |      |
| 2.4.3    | Ordine di focus                            |       |      |
| 2.4.6    | Intestazioni ed etichette                  |       |      |
| 2.4.7    | Focus visibile                             |       |      |
| 2.4.11   | Focus non nascosto (2.2)                   |       |      |
| 2.5.8    | Dimensione target (minimo, 2.2)            |       |      |
| 3.1.1    | Lingua della pagina                        |       |      |
| 3.3.1    | Identificazione errori                     |       |      |
| 3.3.7    | Redundant entry (2.2)                      |       |      |
| 4.1.2    | Nome, ruolo, valore                        |       |      |
| 4.1.3    | Messaggi di stato                          |       |      |

## Non conformità riscontrate

- **[#]** *Pagina/componente* — descrizione del problema — criterio WCAG — severità —
  pianificata risoluzione (issue #NNN, target sprint).

## Sintesi

- **Pagine testate:** N
- **Non conformità critiche:** N
- **Non conformità serie:** N
- **Azioni correttive aperte:** link agli issue

## Firma

- Auditor: ____________________
- Data: YYYY-MM-DD
