# MARKETING-RULES.md — Forbidden Words Contract

**Version:** 1.0
**Authority:** canonical — this file is the contract for the CI anti-word scanner (Phase 1 plan-06).
**Scope anchor:** `.planning/legal/scope.md`

## 1. Purpose

Pesa is a **generic nutritional carb calculator**. Any language that suggests medical, diagnostic, therapeutic, curative, preventive, or "medical device"-style claims is **forbidden** in: source code (`web/src/**/*.{ts,tsx,js,jsx}`, `internal/**/*.go`, `cmd/**/*.go`), translation files (`web/src/i18n/**/*.json`), marketing copy (`README.md`, `.planning/legal/scope/**`, any future `docs/marketing/**`), user-facing release notes (`CHANGELOG.md`). Commit messages are a soft check only.

This file defines **what is forbidden, which regex patterns express it, which severity applies, and what the scanner contract looks like**. Phase 1 plan-06 implements the scanner as pure engineering work: no interpretation decisions at scan time.

The rationale is threefold:
1. **GDPR Art. 9 perimeter exclusion** — no language suggesting health-data processing, since Pesa does not process Art. 9 data by design.
2. **MDR Reg. UE 2017/745 exclusion** — no "intended purpose" claim that could qualify the software as a medical device.
3. **Marketing clarity** — positioning stays neutrally nutritional, target audience is "anyone counting carbs", not a disease cohort.

## 2. Forbidden Words (Italian)

### Cluster A — Dati sanitari / clinici (HARD, no exceptions)

| Pattern | Rationale | Cluster |
|---------|-----------|---------|
| `glicemia`, `glicemic[oaie]` | dato sanitario Art. 9 GDPR; indice glicemico usato come clinical claim | medical |
| `insulin[ae]` | Art. 9 GDPR, farmacologico | medical |
| `bolo`, `boli` | termine terapeutico insulinico | therapeutic |
| `dose`, `dosi`, `dosa` | farmacologico | therapeutic |
| `dosaggio`, `dosaggi` | farmacologico | therapeutic |
| `I:CHO` | parametro insulinico | medical |
| `FSI` | fattore sensibilità insulinica | medical |
| `CGM` | dispositivo medico | medical |
| `microinfusor[ei]` | dispositivo medico | medical |
| `pompa insulinica` / `pompa insulin*` | dispositivo medico | medical |
| `diabete`, `diabetic[oaie]` | positioning forbidden | anti-positioning |
| `prescriv.*` | solo medico può prescrivere | therapeutic |
| `dispositivo medico` | MDR trigger | medical |
| `clinicamente validat.*` | medical claim | medical |
| `pattern glicemic[oi]` | clinical monitoring | medical |

### Cluster B — Claim terapeutici / diagnostici / curativi (HARD)

| Pattern | Rationale | Cluster |
|---------|-----------|---------|
| `cura`, `cure` (IT) | claim curativo | therapeutic |
| `curare` | claim curativo | therapeutic |
| `guarigion.*` | claim curativo | therapeutic |
| `terapi[ae]` | claim terapeutico | therapeutic |
| `terapeutic[oaie]` | claim terapeutico | therapeutic |
| `diagnos[ei]` | claim diagnostico | diagnostic |
| `diagnostic[oaie]` | claim diagnostico | diagnostic |
| `prevenire diabete` | claim preventivo | anti-positioning |
| `prevenzione diabete` | claim preventivo | anti-positioning |
| `trattament[oi]` | claim terapeutico (salvo "trattamento dati" → allow-list) | therapeutic |

### Cluster C — Claim di precisione/autorevolezza medica (SOFT)

| Pattern | Rationale | Cluster |
|---------|-----------|---------|
| `preciso` | evitare claim di precisione medica; preferire "accurato entro i limiti della fonte" | marketing-ambiguity |
| `medico` | tranne "parere medico" in contesti di esclusione (allow-list) | marketing-ambiguity |

## 3. Forbidden Words (English)

### Cluster A — Health / clinical data (HARD)

| Pattern | Rationale |
|---------|-----------|
| `blood glucose` | Art. 9 |
| `insulin` | Art. 9 |
| `bolus` | therapeutic |
| `dose`, `doses` | pharmacological |
| `dosage` | pharmacological |
| `carb ratio` | insulin parameter |
| `ISF` | insulin sensitivity factor |
| `CGM` | medical device |
| `insulin pump` | medical device |
| `diabetes`, `diabetic`, `diabetics` | positioning forbidden |
| `prescribe`, `prescribing`, `prescribed` | clinical only |
| `medical device` | MDR |
| `clinically validated` / `clinically validat.*` | medical claim |

### Cluster B — Therapeutic / diagnostic / curative (HARD)

| Pattern | Rationale |
|---------|-----------|
| `cure`, `cures`, `cured` | curative claim |
| `therapy`, `therapies`, `therapeutic`, `therapeutics` | therapeutic |
| `diagnose`, `diagnoses`, `diagnosis`, `diagnostic` | diagnostic |
| `prevent diabetes` | preventive claim |
| `treatment`, `treatments` | therapeutic (except "data treatment" → allow-list) |

### Cluster C — Precision / medical authority (SOFT)

| Pattern | Rationale |
|---------|-----------|
| `precise` | evita claim di precisione medica |
| `medical` | tranne in contesto di esclusione |

## 4. Regex Patterns

The scanner parses this fenced block line-by-line. Each line is a single regex (POSIX ERE compatible when the `(?i)` prefix is honored by the engine; Go `regexp` via `(?i)` works). Trailing `# comment` is optional.

```
(?i)\bglicemi[aeo]\b
(?i)\bglicemic[oaie]?\b
(?i)\binsulin[ae]?\b
(?i)\bbol[oi]\b
(?i)\bdos[aei]\b
(?i)\bdosaggi[oi]\b
(?i)\bI:CHO\b
(?i)\bFSI\b
(?i)\bCGM\b
(?i)\bmicroinfusor[ei]\b
(?i)\bpompa\s+insulin\w*\b
(?i)\bdiabet[eicoa]+\b
(?i)\bprescriv\w*\b
(?i)\bdispositivo\s+medico\b
(?i)\bclinicamente\s+validat\w*\b
(?i)\bpattern\s+glicemic[oi]\b
(?i)\bcur[ae]\b
(?i)\bcurare\b
(?i)\bguarigion\w*\b
(?i)\bterapi[ae]\b
(?i)\bterapeutic\w*\b
(?i)\bdiagnos[ei]\b
(?i)\bdiagnostic\w*\b
(?i)\bprevenire\s+diabete\b
(?i)\bprevenzione\s+diabete\b
(?i)\btrattament[oi]\b
(?i)\bpreciso\b
(?i)\bmedico\b
(?i)\bblood\s+glucose\b
(?i)\binsulin\b
(?i)\bbolus\b
(?i)\bdose[s]?\b
(?i)\bdosage\b
(?i)\bcarb\s+ratio\b
(?i)\bISF\b
(?i)\binsulin\s+pump\b
(?i)\bdiabet(es|ic|ics)\b
(?i)\bprescrib\w+\b
(?i)\bmedical\s+device\b
(?i)\bclinically\s+validat\w*\b
(?i)\bcure[sd]?\b
(?i)\btherap(y|ies|eutic|eutics)\b
(?i)\bdiagnos(e|es|is|tic)\b
(?i)\bprevent\s+diabetes\b
(?i)\btreatment[s]?\b
(?i)\bprecise\b
(?i)\bmedical\b
```

**Soft-severity subset** (Cluster C — SOFT: emits warning, does not fail build):

```
(?i)\bpreciso\b
(?i)\bmedico\b
(?i)\bprecise\b
(?i)\bmedical\b
```

All other patterns are HARD.

## 5. Scanner Contract

The scanner implemented in **Phase 1 plan-06** MUST satisfy the following contract.

### 5.1 Name and invocation

- **Name:** `scripts/check-forbidden-words.sh` or equivalent `cmd/check-forbidden-words/main.go` (implementer's choice; the shell wrapper is acceptable provided it can be called from CI).
- **Invocation:** no required arguments; scans the repo from its root.

### 5.2 Input

Reads regex patterns from the `## 4. Regex Patterns` section of this very file by:

1. locating the H2 heading `## 4. Regex Patterns` (or `## 4.`)
2. reading the **first fenced block** that follows (between the first `` ``` `` and the matching `` ``` ``)
3. treating every non-empty, non-comment line as a pattern (comments are `#` to end-of-line, after stripping leading/trailing whitespace).

A **second fenced block** in §4 (under the "Soft-severity subset" sub-heading) defines which patterns are SOFT. Patterns not in that subset are HARD.

### 5.3 File globs scanned (in order)

- `web/src/**/*.{ts,tsx,js,jsx}`
- `web/src/i18n/**/*.json`
- `internal/**/*.go`
- `cmd/**/*.go`
- `docs/**/*.md` (excluding `docs/legal/**` and `docs/reference/**`)
- `README.md`
- `CHANGELOG.md`
- `.planning/legal/scope/*.md`

### 5.4 Explicitly NOT scanned (allow-list paths)

- `MARKETING-RULES.md` itself (contains the words by definition)
- `.planning/**` except `.planning/legal/scope/`
- `LICENSE`, `LICENSE.*`
- `*.lock`, `*.sum`, `go.mod`, `package.json`, `pnpm-lock.yaml`, `package-lock.json`, `yarn.lock`
- `tests/forbidden-words/**` (intentional negative test fixtures)

### 5.5 Severity tiers

- **HARD** (exit 1): all patterns in Clusters A and B (IT + EN).
- **SOFT** (exit 0 + warning on stderr): all patterns in Cluster C (IT + EN).

### 5.6 Inline allow-list (escape hatch)

A source line may suppress **one match of one specific pattern** when it contains the marker comment:

- `// marketing-rules: allow <pattern-id>` (for `.ts`, `.tsx`, `.js`, `.jsx`, `.go`)
- `{/* marketing-rules: allow <pattern-id> */}` (for JSX text nodes)
- `<!-- marketing-rules: allow <pattern-id> -->` (for `.md` sources permitted to quote forbidden terms in their natural legal/educational function, e.g. scope exclusion disclaimers)

Rules:
- The marker must be **on the same line as the match**, OR on the line immediately preceding it.
- Every use MUST include a human-readable rationale as trailing free text, e.g.: `// marketing-rules: allow medico — quoted in disclaimer "non sostituisce il parere medico"`.
- The scanner logs every allow-list suppression with `file:line`, pattern-id, and rationale to the CI job log for periodic review.
- A line with no rationale (`// marketing-rules: allow medico` alone, no trailing text after `—`) is treated as **invalid** and the match is NOT suppressed (fail-safe).

### 5.7 Context-aware automatic exceptions

Compiled into the scanner as fixed whitelist (no configuration):

1. **`trattamento dati` / `trattamento dei dati` / `data treatment`** — legal term, not therapy → automatic pass for the `trattament[oi]` / `treatment[s]?` pattern when the 3 tokens to the right (IT) or the 2 tokens to the right (EN) contain `dat[io]` / `data` / `personali`.
2. **`parere medico` / `medical opinion`** — explicit exclusion disclaimer context → automatic pass for the `medico` / `medical` SOFT pattern when preceded by `parere` / `opinion of a` within 3 tokens to the left.
3. **`not a medical device` / `non è un dispositivo medico`** — MDR exclusion context → automatic pass for the `medical device` / `dispositivo medico` pattern when preceded by a negation (`not`, `non`, `never`) within 3 tokens.

These exceptions apply **only** to the patterns listed; they do not generalize.

### 5.8 Exit codes

- `0` — no HARD matches (SOFT warnings allowed and reported)
- `1` — at least one HARD match (build fails)
- `2` — internal error (file unreadable, regex parse fail, glob error, pattern block not found)

### 5.9 Output format

For each match, one line on stdout:

```
LEVEL\tFILE:LINE:COL\tPATTERN\tMATCHED_TEXT\tCONTEXT
```

- Tab-separated, UTF-8, one match per line.
- `LEVEL` ∈ `{HARD, SOFT, ALLOWLISTED}`.
- `CONTEXT` = ±20 characters around the match, with newlines replaced by spaces.
- CI greps for `^HARD\t` to decide fail.

SOFT warnings also mirror to stderr prefixed by `[marketing-rules warning] `.

## 6. Positive Language Guidelines

Use these phrasings in UI copy and documentation:

- instead of "preciso" → **"accurato entro i limiti della fonte"**
- instead of "per diabetici" → **"per chi conta i carboidrati"**
- instead of "terapia nutrizionale" → **"regime alimentare"**
- instead of "dose di carboidrati" → **"quantità di carboidrati"**
- instead of "cura nutrizionale" → **"alimentazione consapevole"**
- instead of "medico" as authority claim → **source name** (e.g. "secondo i dati CREA")
- instead of "clinically validated" → **"based on published Italian food composition tables"**
- instead of "therapeutic diet" → **"specific dietary regimen"**
- instead of "treatment" (as verb) → **"processing"** (for data) or **"management"** (for generic workflow)

## 7. Maintenance

- Additions to the patterns in §4 MUST be accompanied by: (a) a rationale added to §2 or §3, (b) a commit at repo root touching this file, (c) a negative test fixture in `.planning/legal/marketing-rules-fixtures.md` (and eventually `tests/forbidden-words/**` when the scanner ships).
- Removals from the patterns require an issue/PR with justification; must be reviewed by the titolare.
- The scanner in Phase 1 plan-06 MUST fail CI if `MARKETING-RULES.md` is modified in a PR but the pattern block cannot be parsed (regex syntax error, missing fenced block, etc.) — this "parser smoke-test" ensures the contract remains machine-consumable.
- Version bump (`v1.0` → `v1.1` for additions, `v2.0` for breaking changes to scanner contract) is required on every substantive edit; recorded in the version history footer.

## 8. Version history

- **v1.0** — first emission in Phase 0 plan-02. 47 HARD + 4 SOFT patterns covering IT and EN clusters A, B, C.
