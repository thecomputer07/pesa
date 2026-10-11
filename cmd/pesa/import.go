package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/thecomputer07/pesa/internal/importer"
)

// runImport dispatches `pesa import <target> <path> [flags]`.
func runImport(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import <crea|crea-live|bda|off|off-dump|ciqual|usda> [...]")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "crea":
		runImportCREA(rest)
	case "bda":
		runImportBDA(rest)
	case "off":
		runImportOFF(rest)
	case "off-dump":
		runImportOFFDump(rest)
	case "ciqual":
		runImportCiqual(rest)
	case "usda":
		runImportUSDA(rest)
	case "crea-live":
		runImportCREALive(rest)
	case "wiki-images":
		runEnrichWikiImages(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown import target: %s\n", sub)
		os.Exit(2)
	}
}

func runImportCREALive(args []string) {
	fs := flag.NewFlagSet("import crea-live", flag.ExitOnError)
	limit := fs.Int("limit", 0, "max alimenti (0 = tutti)")
	rateMs := fs.Int("rate-ms", 2000, "delay tra richieste in ms")
	dryRun := fs.Bool("dry-run", false, "parse e check senza scrivere")
	missing := fs.Bool("missing", false, "solo gli alimenti dell'indice non ancora in catalogo")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()
	opts := importer.CREALiveOpts{
		OnlyMissing: *missing,
		ImportOpts:  importer.ImportOpts{SourceCode: "crea", Basis: importer.BasisEdible, DryRun: *dryRun},
		Limit:       *limit,
		RateDelay:   time.Duration(*rateMs) * time.Millisecond,
	}
	report, err := importer.ImportCREALive(context.Background(), db, opts)
	if err != nil {
		slog.Error("import crea-live", "err", err)
		os.Exit(1)
	}
	slog.Info("import crea-live complete",
		"total", report.TotalRows, "inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods, "unchanged", report.Unchanged,
		"quarantined", report.Quarantined, "errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for i, e := range report.Errors {
			if i >= 20 {
				slog.Warn("... more errors truncated", "total", len(report.Errors))
				break
			}
			slog.Warn("row error", "detail", e)
		}
	}
}

func runEnrichWikiImages(args []string) {
	fs := flag.NewFlagSet("import wiki-images", flag.ExitOnError)
	limit := fs.Int("limit", 0, "max foods (0 = tutti)")
	lang := fs.String("lang", "it", "wikipedia lang code")
	rateMs := fs.Int("rate-ms", 250, "delay tra richieste in ms")
	source := fs.String("source", "", "limita a food_sources.code (es. crea)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()
	opts := importer.WikiImageOpts{
		Limit:      *limit,
		Lang:       *lang,
		RateDelay:  time.Duration(*rateMs) * time.Millisecond,
		OnlySource: *source,
	}
	report, err := importer.EnrichImagesFromWikipedia(context.Background(), db, opts)
	if err != nil {
		slog.Error("wiki-images", "err", err)
		os.Exit(1)
	}
	slog.Info("wiki-images complete",
		"total", report.Total, "found", report.Found, "missing", report.Missing,
		"errors", len(report.Errors))
}

func runImportOFF(args []string) {
	fs := flag.NewFlagSet("import off", flag.ExitOnError)
	country := fs.String("country", "italy", "countries_tags_en filter")
	limit := fs.Int("limit", 5000, "max prodotti totali (0 = no cap)")
	pageSize := fs.Int("page-size", 100, "prodotti per pagina")
	rateMs := fs.Int("rate-ms", 1000, "delay tra richieste API in millisecondi")
	minCarb := fs.Bool("min-carb", true, "scarta prodotti senza carbohydrates_100g")
	startPage := fs.Int("start-page", 1, "riprendi da questa pagina (dopo un'interruzione)")
	byScans := fs.Bool("by-scans", false, "ordina per numero di scansioni (i piu' consumati) via API v2; rate-ms consigliato >= 7000")
	dryRun := fs.Bool("dry-run", false, "parse e check senza scrivere")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.OFFOpts{
		ImportOpts: importer.ImportOpts{
			SourceCode: "off",
			Basis:      importer.BasisEdible,
			DryRun:     *dryRun,
		},
		Country:   *country,
		Limit:     *limit,
		PageSize:  *pageSize,
		RateDelay: time.Duration(*rateMs) * time.Millisecond,
		MinCarb:   *minCarb,
		ByScans:   *byScans,
		StartPage: *startPage,
	}
	if *byScans && *rateMs == 1000 {
		opts.RateDelay = 0 // let the importer pick the v2 rate limit
	}
	report, err := importer.ImportOFF(context.Background(), db, opts)
	if err != nil {
		slog.Error("import off", "err", err)
		os.Exit(1)
	}
	slog.Info("import off complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for i, e := range report.Errors {
			if i >= 20 {
				slog.Warn("... more errors truncated", "total", len(report.Errors))
				break
			}
			slog.Warn("row error", "detail", e)
		}
	}
}

func runImportCREA(args []string) {
	fs := flag.NewFlagSet("import crea", flag.ExitOnError)
	basis := fs.String("basis", "edible", "edible | as_purchased")
	dryRun := fs.Bool("dry-run", false, "parse and check without writing")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import crea <csv> [--basis] [--dry-run]")
		os.Exit(2)
	}
	path := fs.Arg(0)

	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.ImportOpts{
		SourceCode: "crea",
		Basis:      importer.Basis(*basis),
		DryRun:     *dryRun,
	}
	report, err := importer.ImportCSV(context.Background(), db, path, opts)
	if err != nil {
		slog.Error("import crea", "err", err)
		os.Exit(1)
	}
	slog.Info("import crea complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for _, e := range report.Errors {
			slog.Warn("row error", "detail", e)
		}
		os.Exit(1)
	}
}

func runImportBDA(args []string) {
	fs := flag.NewFlagSet("import bda", flag.ExitOnError)
	format := fs.String("format", "", "csv | xlsx (default: auto)")
	basis := fs.String("basis", "edible", "edible | as_purchased")
	dryRun := fs.Bool("dry-run", false, "parse and check without writing")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import bda <path> [--format] [--basis] [--dry-run]")
		os.Exit(2)
	}
	path := fs.Arg(0)
	_ = format // reserved for future explicit override; auto-detect covers all cases now

	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.ImportOpts{
		SourceCode: "bda-ieo",
		Basis:      importer.Basis(*basis),
		DryRun:     *dryRun,
	}
	report, err := importer.ImportBDA(context.Background(), db, path, opts)
	if err != nil {
		slog.Error("import bda", "err", err)
		os.Exit(1)
	}
	slog.Info("import bda complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"superseded_by_crea", report.SupersededByCREA,
		"errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for _, e := range report.Errors {
			slog.Warn("row error", "detail", e)
		}
		os.Exit(1)
	}
}

func runImportCiqual(args []string) {
	fs := flag.NewFlagSet("import ciqual", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "parse and check without writing")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import ciqual <dir con alim_*.xml e compo_*.xml> [--dry-run]")
		os.Exit(2)
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.ImportOpts{SourceCode: "ciqual", Basis: importer.BasisEdible, DryRun: *dryRun}
	report, err := importer.ImportCiqual(context.Background(), db, fs.Arg(0), opts)
	if err != nil {
		slog.Error("import ciqual", "err", err)
		os.Exit(1)
	}
	slog.Info("import ciqual complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for _, e := range report.Errors {
			slog.Warn("row error", "detail", e)
		}
		os.Exit(1)
	}
}
func runImportUSDA(args []string) {
	fs := flag.NewFlagSet("import usda", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "parse and check without writing")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import usda <dir con food.csv e food_nutrient.csv (FNDDS)> [--dry-run]")
		os.Exit(2)
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.ImportOpts{SourceCode: "usda", Basis: importer.BasisEdible, DryRun: *dryRun}
	report, err := importer.ImportUSDA(context.Background(), db, fs.Arg(0), opts)
	if err != nil {
		slog.Error("import usda", "err", err)
		os.Exit(1)
	}
	slog.Info("import usda complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"errors", len(report.Errors))
	if len(report.Errors) > 0 {
		for _, e := range report.Errors {
			slog.Warn("row error", "detail", e)
		}
		os.Exit(1)
	}
}

func runImportOFFDump(args []string) {
	fs := flag.NewFlagSet("import off-dump", flag.ExitOnError)
	country := fs.String("country", "Italy", "countries_en deve contenere questo paese")
	top := fs.Int("top", 10000, "i N prodotti piu' scansionati (0 = tutti)")
	dryRun := fs.Bool("dry-run", false, "parse e check senza scrivere")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: pesa import off-dump <en.openfoodfacts.org.products.csv[.gz] | -> [--top N] [--country Italy]")
		os.Exit(2)
	}
	var in io.Reader = os.Stdin
	if p := fs.Arg(0); p != "-" {
		f, err := os.Open(p)
		if err != nil {
			slog.Error("import off-dump", "err", err)
			os.Exit(1)
		}
		defer func() { _ = f.Close() }()
		in = f
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()

	opts := importer.OFFDumpOpts{
		OFFOpts: importer.OFFOpts{
			ImportOpts: importer.ImportOpts{SourceCode: "off", Basis: importer.BasisEdible, DryRun: *dryRun},
			MinCarb:    true,
		},
		Country: *country,
		Top:     *top,
	}
	report, err := importer.ImportOFFDump(context.Background(), db, in, opts)
	if err != nil {
		slog.Error("import off-dump", "err", err)
		os.Exit(1)
	}
	slog.Info("import off-dump complete",
		"total", report.TotalRows,
		"inserted", report.InsertedFoods,
		"updated", report.UpdatedFoods,
		"unchanged", report.Unchanged,
		"quarantined", report.Quarantined,
		"errors", len(report.Errors))
}
