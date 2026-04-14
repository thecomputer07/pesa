package export

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// AccountJSON is the top-level "account.json" bundle entry. Column allowlist
// enforces T-06-04: no password_hash, reset tokens, session ids.
type AccountJSON struct {
	Email           string     `json:"email"`
	Lang            string     `json:"lang"`
	CreatedAt       time.Time  `json:"created_at"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time `json:"last_login_at,omitempty"`
}

// Pseudonym returns SHA-256(email + salt)[:16] (Phase 1 logger pseudonym).
func Pseudonym(email string, salt []byte) string {
	h := sha256.New()
	h.Write([]byte(email))
	h.Write(salt)
	return trimPseudonym(hex.EncodeToString(h.Sum(nil)))
}

// BuildAccountJSON loads the account row for userID and marshals the
// allowlisted fields to JSON. Never includes credential material.
func BuildAccountJSON(db *sql.DB, userID int64) ([]byte, error) {
	var a AccountJSON
	var verifiedAt, lastLogin sql.NullTime
	var lang sql.NullString
	err := db.QueryRow(
		`SELECT email, COALESCE(lang,''), created_at, email_verified_at, last_login_at
		 FROM users WHERE id = $1`, userID).
		Scan(&a.Email, &lang, &a.CreatedAt, &verifiedAt, &lastLogin)
	if err != nil {
		return nil, err
	}
	a.Lang = lang.String
	if verifiedAt.Valid {
		t := verifiedAt.Time
		a.EmailVerifiedAt = &t
	}
	if lastLogin.Valid {
		t := lastLogin.Time
		a.LastLoginAt = &t
	}
	return json.MarshalIndent(a, "", "  ")
}

// WriteCSV writes RFC 4180-conformant CSV. The first row is the supplied
// header slice; subsequent rows come from rows (each []string, equal length).
// Deterministic output: callers must pre-order their rows.
func WriteCSV(w io.Writer, header []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(header); err != nil {
		return err
	}
	for _, r := range rows {
		if len(r) != len(header) {
			return fmt.Errorf("export: row width %d != header %d", len(r), len(header))
		}
		if err := cw.Write(r); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// TablePart is one CSV entry in the bundle.
type TablePart struct {
	Name   string
	Header []string
	Rows   [][]string
}

// CollectTables runs the per-table allowlist SELECTs for userID and returns
// the ordered TablePart slice. Empty tables yield a header-only CSV — the
// presence of a zero-row file is itself informative to the data subject.
func CollectTables(db *sql.DB, userID int64) ([]TablePart, error) {
	parts := []TablePart{}

	// meals
	header := []string{"id", "eaten_at", "slot", "notes", "updated_at"}
	rows, err := selectRows(db,
		`SELECT id, eaten_at, slot, COALESCE(notes,''), updated_at FROM meals WHERE user_id=$1 ORDER BY id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"meals.csv", header, rows})

	// meal_items
	header = []string{"id", "meal_id", "food_id", "grams", "recipe_client_id", "updated_at"}
	rows, err = selectRows(db,
		`SELECT mi.id, mi.meal_id, COALESCE(mi.food_id,0), mi.grams,
		        COALESCE(mi.recipe_client_id,''), mi.updated_at
		 FROM meal_items mi JOIN meals m ON m.id = mi.meal_id
		 WHERE m.user_id = $1 ORDER BY mi.id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"meal_items.csv", header, rows})

	// recipes
	header = []string{"id", "client_id", "name", "servings", "total_grams", "notes", "updated_at"}
	rows, err = selectRows(db,
		`SELECT id, client_id, name, servings, COALESCE(total_grams,0),
		        COALESCE(notes,''), updated_at
		 FROM recipes WHERE user_id=$1 ORDER BY id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"recipes.csv", header, rows})

	// recipe_items
	header = []string{"id", "recipe_id", "food_id", "grams", "updated_at"}
	rows, err = selectRows(db,
		`SELECT ri.id, ri.recipe_id, ri.food_id, ri.grams, ri.updated_at
		 FROM recipe_items ri JOIN recipes r ON r.id = ri.recipe_id
		 WHERE r.user_id = $1 ORDER BY ri.id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"recipe_items.csv", header, rows})

	// meal_templates
	header = []string{"id", "name", "slot", "payload_json", "updated_at"}
	rows, err = selectRows(db,
		`SELECT id, name, slot, COALESCE(payload_json,''), updated_at
		 FROM meal_templates WHERE user_id=$1 ORDER BY id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"meal_templates.csv", header, rows})

	// favorites
	header = []string{"id", "food_id", "created_at"}
	rows, err = selectRows(db,
		`SELECT id, food_id, created_at FROM favorites WHERE user_id=$1 ORDER BY id`,
		userID, len(header))
	if err != nil {
		return nil, err
	}
	parts = append(parts, TablePart{"favorites.csv", header, rows})

	// sessions metadata (no tokens)
	header = []string{"token_prefix", "created_at", "expires_at"}
	rows, err = selectRows(db,
		`SELECT SUBSTRING(token,1,8) AS token_prefix, created_at, expires_at
		 FROM sessions WHERE user_id=$1 ORDER BY expires_at DESC`,
		userID, len(header))
	if err != nil {
		// sessions table may have different schema in some environments;
		// swallow column-missing errors as empty.
		parts = append(parts, TablePart{"sessions.csv", header, [][]string{}})
	} else {
		parts = append(parts, TablePart{"sessions.csv", header, rows})
	}

	return parts, nil
}

func selectRows(db *sql.DB, query string, arg interface{}, width int) ([][]string, error) {
	rs, err := db.Query(query, arg)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rs.Close() }()
	out := [][]string{}
	for rs.Next() {
		cols := make([]any, width)
		ptrs := make([]any, width)
		for i := range cols {
			ptrs[i] = &cols[i]
		}
		if err := rs.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, width)
		for i, c := range cols {
			row[i] = coerce(c)
		}
		out = append(out, row)
	}
	return out, rs.Err()
}

func coerce(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case []byte:
		return string(x)
	case string:
		return x
	case int64:
		return fmt.Sprintf("%d", x)
	case float64:
		return fmt.Sprintf("%g", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case time.Time:
		return x.UTC().Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", x)
	}
}

// ReadmeMarkdown returns the stable README accompanying the bundle.
func ReadmeMarkdown() []byte {
	return []byte(`# Pesa — Personal Data Export

Questo archivio contiene i tuoi dati personali presenti su Pesa al momento
dell'export, ai sensi degli art. 15 e 20 del GDPR.

## Contenuto

- ` + "`account.json`" + ` — anagrafica account (email, lingua, date di creazione e verifica)
- ` + "`meals.csv`" + ` — pasti registrati
- ` + "`meal_items.csv`" + ` — alimenti all'interno dei pasti
- ` + "`recipes.csv`" + ` — ricette create
- ` + "`recipe_items.csv`" + ` — ingredienti delle ricette
- ` + "`meal_templates.csv`" + ` — modelli di pasto salvati
- ` + "`favorites.csv`" + ` — alimenti preferiti
- ` + "`sessions.csv`" + ` — metadati sessioni (solo prefisso anonimo, mai il token)

Password, token di sessione, token di reset e CSRF **non** sono inclusi.

## Licenza dei dati

Questi dati sono tuoi. Puoi riutilizzarli, importarli altrove o conservarli
come copia personale.

Contatto: privacy@pesadia.com
`)
}
