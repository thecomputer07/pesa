package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/thecomputer07/pesa/internal/scraper/crea"
	"github.com/thecomputer07/pesa/internal/scraper/httpclient"
	"github.com/thecomputer07/pesa/internal/scraper/robots"
)

// runScrape dispatches `pesa scrape <target>` subcommands.
func runScrape(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: pesa scrape <crea> [flags]")
		os.Exit(2)
	}
	sub := args[0]
	rest := args[1:]
	switch sub {
	case "crea":
		runScrapeCREA(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown scrape target: %s\n", sub)
		os.Exit(2)
	}
}

func runScrapeCREA(args []string) {
	fs := flag.NewFlagSet("scrape crea", flag.ExitOnError)
	output := fs.String("output", "data/crea/"+time.Now().UTC().Format("2006-01-02")+".csv", "CSV output path")
	cacheDir := fs.String("cache", ".cache/crea", "HTML cache dir (empty to disable)")
	rate := fs.Duration("rate", 2*time.Second, "min interval between requests")
	limit := fs.Int("limit", 0, "max detail pages (0 = all)")
	indexURL := fs.String("index-url", "{{CREA_INDEX_URL}}", "index page URL")
	userAgent := fs.String("user-agent", "Pesa/dev (+https://github.com/thecomputer07/pesa; contact: lorenzo.dimare@icloud.com)", "HTTP User-Agent")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	ctx := context.Background()
	client := httpclient.New(*userAgent, *rate, *cacheDir)
	checker := robots.NewChecker(*userAgent)

	if ok, err := checker.Allowed(ctx, *indexURL); err != nil {
		slog.Error("robots check failed", "err", err)
		os.Exit(1)
	} else if !ok {
		slog.Error("robots.txt disallows index", "url", *indexURL)
		os.Exit(1)
	}

	indexBody, _, err := client.Get(ctx, *indexURL)
	if err != nil {
		slog.Error("fetch index", "err", err)
		os.Exit(1)
	}
	urls, err := crea.ParseIndexHTML(indexBody, *indexURL)
	if err != nil {
		slog.Error("parse index", "err", err)
		os.Exit(1)
	}
	if *limit > 0 && len(urls) > *limit {
		urls = urls[:*limit]
	}

	if err := os.MkdirAll(dirOf(*output), 0o755); err != nil {
		slog.Error("mkdir output", "err", err)
		os.Exit(1)
	}
	outF, err := os.Create(*output)
	if err != nil {
		slog.Error("create output", "err", err)
		os.Exit(1)
	}
	defer func() { _ = outF.Close() }()
	w := csv.NewWriter(outF)
	defer w.Flush()
	_ = w.Write([]string{
		"external_id", "url", "name_it", "category", "edible_pct", "kcal",
		"protein_g", "fat_g", "carb_g", "sugar_g", "fiber_g", "water_g",
		"ash_g", "alcohol_g", "salt_g", "content_hash", "fetched_at",
	})

	errF, _ := os.Create(*output + ".errors.csv")
	defer func() { _ = errF.Close() }()
	errW := csv.NewWriter(errF)
	defer errW.Flush()
	_ = errW.Write([]string{"url", "error", "timestamp"})

	var ok, failed, hits int
	for _, u := range urls {
		if allowed, rerr := checker.Allowed(ctx, u); rerr != nil || !allowed {
			slog.Warn("skip disallowed", "url", u)
			_ = errW.Write([]string{u, "robots_disallow", time.Now().UTC().Format(time.RFC3339)})
			failed++
			continue
		}
		start := time.Now()
		body, cached, ferr := client.Get(ctx, u)
		if ferr != nil {
			_ = errW.Write([]string{u, ferr.Error(), time.Now().UTC().Format(time.RFC3339)})
			failed++
			continue
		}
		if cached {
			hits++
		}
		rec, perr := crea.ParseDetailHTML(body, u)
		if perr != nil {
			_ = errW.Write([]string{u, perr.Error(), time.Now().UTC().Format(time.RFC3339)})
			failed++
			continue
		}
		_ = w.Write(recordToCSV(rec))
		slog.Info("crea fetch", "url", u, "cache_hit", cached, "duration_ms", time.Since(start).Milliseconds())
		ok++
	}
	slog.Info("crea scrape complete", "total", len(urls), "ok", ok, "errors", failed, "cache_hits", hits)
}

func recordToCSV(r *crea.CREARecord) []string {
	f := func(p *float64) string {
		if p == nil {
			return ""
		}
		return strconv.FormatFloat(*p, 'f', -1, 64)
	}
	return []string{
		r.ExternalID, r.URL, r.NameIT, r.Category,
		f(r.EdiblePct), f(r.KcalPer100g),
		f(r.ProteinG), f(r.FatG), f(r.CarbG), f(r.SugarG),
		f(r.FiberG), f(r.WaterG), f(r.AshG), f(r.AlcoholG), f(r.SaltG),
		r.ContentHash, r.FetchedAt,
	}
}

func dirOf(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}
