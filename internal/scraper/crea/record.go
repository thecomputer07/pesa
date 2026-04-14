// Package crea holds the CREA-specific scraper logic: record shape + HTML parsers.
package crea

// CREARecord is the parsed payload of a single alimentinutrizione.it detail page.
//
// All nutritional values are "per 100 g of edible portion", matching CREA's
// published basis. See internal/scraper/README.md.
type CREARecord struct {
	ExternalID  string
	URL         string
	NameIT      string
	Category    string
	EdiblePct   *float64
	KcalPer100g *float64
	ProteinG    *float64
	FatG        *float64
	CarbG       *float64
	SugarG      *float64
	FiberG      *float64
	WaterG      *float64
	AshG        *float64
	AlcoholG    *float64
	SaltG       *float64

	// ExtraFields collects the remaining labeled rows (micronutrients, etc.).
	ExtraFields map[string]string

	FetchedAt   string // RFC 3339
	ContentHash string // sha256 hex of source HTML (64 chars)
}
