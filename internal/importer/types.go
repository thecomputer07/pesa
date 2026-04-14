// Package importer ingests CSV/XLSX food data into the `foods` table
// with normalization, sanity checks, and a quarantine path.
package importer

import "time"

// FoodRow is the common in-memory row shape handled by the import pipeline.
type FoodRow struct {
	ExternalID  string
	URL         string
	NameIT      string
	Category    string
	EdiblePct   float64
	Kcal        *float64
	ProteinG    *float64
	FatG        *float64
	CarbG       *float64
	SugarG      *float64
	FiberG      *float64
	WaterG      *float64
	AshG        *float64
	AlcoholG    *float64
	SaltG       *float64
	ContentHash string
	FetchedAt   time.Time
	RawJSON     []byte
}

// Basis indicates whether values are reported per-100g-edible or per-100g-as-purchased.
type Basis string

const (
	BasisEdible      Basis = "edible"
	BasisAsPurchased Basis = "as_purchased"
)

// SanityResult reports macro/kcal balance verdicts.
type SanityResult struct {
	OK       bool
	Reasons  []string
	Severity string // "warn" | "error"
}

// ImportOpts configures a single ImportCSV / ImportBDA invocation.
type ImportOpts struct {
	SourceCode string
	Basis      Basis
	DryRun     bool
}

// ImportReport aggregates the outcome of an import run.
type ImportReport struct {
	TotalRows         int
	InsertedFoods     int
	UpdatedFoods      int
	Unchanged         int
	Quarantined       int
	SupersededByCREA  int
	Errors            []string
}
