package importer

import (
	"context"
	"database/sql"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Italian function words stripped by NormalizeName.
var itStopwords = map[string]struct{}{
	"di": {}, "del": {}, "della": {}, "dello": {}, "dei": {}, "degli": {},
	"delle": {}, "al": {}, "alla": {}, "allo": {}, "agli": {}, "alle": {},
	"con": {}, "e": {}, "o": {}, "la": {}, "il": {}, "lo": {}, "gli": {},
	"le": {}, "un": {}, "una": {}, "uno": {}, "in": {}, "da": {}, "su": {},
	"per": {}, "a": {},
}

// NormalizeName lowercases, strips combining marks, drops Italian stopwords,
// collapses whitespace — used for CREA/BDA title matching.
func NormalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// NFD → drop combining marks → NFC.
	decomposed := norm.NFD.String(s)
	var b strings.Builder
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	stripped := norm.NFC.String(b.String())

	// Replace non-alphanumeric with space.
	var clean strings.Builder
	for _, r := range stripped {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			clean.WriteRune(r)
		} else {
			clean.WriteRune(' ')
		}
	}
	tokens := strings.Fields(clean.String())
	kept := tokens[:0]
	for _, t := range tokens {
		if _, skip := itStopwords[t]; skip {
			continue
		}
		kept = append(kept, t)
	}
	return strings.Join(kept, " ")
}

// FindCREAMatch returns the food id of a CREA row whose normalized name + state
// equal the inputs, or 0 if none.
//
// Implementation: pg_trgm similarity prefilter against normName (for speed),
// then NormalizeName comparison on result set.
func FindCREAMatch(ctx context.Context, db *sql.DB, normName, state string) (int64, error) {
	if normName == "" {
		return 0, nil
	}

	rows, err := db.QueryContext(ctx, `
		SELECT f.id, f.name_it
		  FROM foods f
		  JOIN food_sources s ON s.id = f.source_id
		 WHERE s.code = 'crea'
		   AND f.state = $1
		   AND similarity(f.name_it, $2) > 0.4
		 ORDER BY similarity(f.name_it, $2) DESC
		 LIMIT 5`,
		state, normName)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return 0, err
		}
		if NormalizeName(name) == normName {
			return id, nil
		}
	}
	return 0, rows.Err()
}
