package importer

import (
	"context"
	"database/sql"
	"fmt"
)

// ImportBDA ingests a BDA IEO dataset (CSV or XLSX), applying sanity checks +
// CREA>BDA supersede precedence. BDA rows matching an existing CREA row
// (NormalizeName + state) are inserted with superseded_by set to the CREA id,
// preserving traceability while excluding the row from search results.
func ImportBDA(ctx context.Context, db *sql.DB, path string, opts ImportOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "bda-ieo"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}

	format := FormatAuto
	rows, parseReport, err := ParseBDA(path, format)
	if err != nil {
		return nil, fmt.Errorf("parse bda: %w", err)
	}

	sourceID, err := lookupSourceID(ctx, db, opts.SourceCode)
	if err != nil {
		return nil, err
	}

	report := &ImportReport{TotalRows: len(rows)}
	if parseReport != nil && parseReport.RowsFailed > 0 {
		report.Errors = append(report.Errors, parseReport.Errors...)
	}

	for _, row := range rows {
		if err := importOneRow(ctx, db, sourceID, row, opts, report, true); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", row.ExternalID, err))
		}
	}
	return report, nil
}
