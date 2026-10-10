package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/thecomputer07/pesa/internal/service/imagecache"
)

// runImages dispatches `pesa images <sub>`.
func runImages(args []string) {
	if len(args) < 1 || args[0] != "localize" {
		fmt.Fprintln(os.Stderr, "usage: pesa images localize [--limit N] [--workers N]")
		os.Exit(2)
	}
	runImagesLocalize(args[1:])
}

// runImagesLocalize downloads external food_images (OFF, Wikimedia) into the
// local image cache and rewrites their url to /static/images/... The CSP
// only allows same-origin images, and serving them ourselves keeps users'
// browsers from contacting third parties. The original URL stays in the
// cache sidecar (original_url).
func runImagesLocalize(args []string) {
	fs := flag.NewFlagSet("images localize", flag.ExitOnError)
	limit := fs.Int("limit", 0, "max immagini (0 = tutte)")
	workers := fs.Int("workers", 2, "download paralleli")
	delay := fs.Duration("delay", 500*time.Millisecond, "pausa per worker tra download (Wikimedia risponde 429 se troppo veloce)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	cfg := mustConfig()
	db := openDB(cfg)
	defer func() { _ = db.Close() }()
	store, err := imagecache.NewStore(cfg.ImageCacheDir)
	if err != nil {
		slog.Error("image store", "err", err)
		os.Exit(1)
	}
	fetcher := imagecache.NewFetcher()
	ctx := context.Background()

	q := `SELECT i.id, i.url, coalesce(i.license, ''), coalesce(i.attribution, '')
	        FROM food_images i
	       WHERE i.is_external AND i.url LIKE 'https://%'
	       ORDER BY i.id`
	if *limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", *limit)
	}
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		slog.Error("query food_images", "err", err)
		os.Exit(1)
	}
	type job struct {
		id                        int64
		url, license, attribution string
	}
	var jobs []job
	for rows.Next() {
		var j job
		if err := rows.Scan(&j.id, &j.url, &j.license, &j.attribution); err != nil {
			slog.Error("scan", "err", err)
			os.Exit(1)
		}
		jobs = append(jobs, j)
	}
	_ = rows.Close()
	slog.Info("images localize start", "total", len(jobs))

	var done, failed atomic.Int64
	ch := make(chan job)
	var wg sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range ch {
				// Be polite to the upstream CDNs, also after a failure.
				time.Sleep(*delay)
				u, err := url.Parse(j.url)
				if err != nil {
					failed.Add(1)
					continue
				}
				var source string
				switch u.Host {
				case "images.openfoodfacts.org", "static.openfoodfacts.org":
					source = "off"
				case "upload.wikimedia.org":
					source = "wikimedia"
				default:
					failed.Add(1)
					continue
				}
				body, ct, err := fetcher.Fetch(ctx, j.url)
				if err != nil {
					failed.Add(1)
					slog.Warn("image fetch", "id", j.id, "err", err)
					continue
				}
				sha, ext, err := store.Put(source, body, ct, imagecache.SidecarMeta{
					Source: source, License: j.license, Attribution: j.attribution,
					OriginalURL: j.url, FetchedAt: time.Now().UTC().Format(time.RFC3339),
				})
				if err != nil {
					failed.Add(1)
					slog.Warn("image store", "id", j.id, "err", err)
					continue
				}
				if _, err := db.ExecContext(ctx,
					`UPDATE food_images SET url = $1, is_external = false WHERE id = $2`,
					"/static/images/"+source+"/"+sha+ext, j.id); err != nil {
					failed.Add(1)
					slog.Warn("image update", "id", j.id, "err", err)
					continue
				}
				if n := done.Add(1); n%250 == 0 {
					slog.Info("images localize progress", "done", n, "failed", failed.Load(), "total", len(jobs))
				}
			}
		}()
	}
	for _, j := range jobs {
		ch <- j
	}
	close(ch)
	wg.Wait()
	slog.Info("images localize complete", "done", done.Load(), "failed", failed.Load(), "total", len(jobs))
}
