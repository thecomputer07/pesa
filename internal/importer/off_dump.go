package importer

import (
	"bufio"
	"compress/gzip"
	"container/heap"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

// OFFDumpOpts controls the import from the Open Food Facts CSV export
// (https://static.openfoodfacts.org/data/en.openfoodfacts.org.products.csv.gz).
// The export has no request limits, unlike the search API (anonymous users
// get 401 past page 10), and carries unique_scans_n for every product.
type OFFDumpOpts struct {
	OFFOpts
	Country string // countries_en must contain it, e.g. "Italy"
	Top     int    // keep the N most scanned products (0 = all matching)
}

// offDumpItem is a candidate kept while streaming: only what we import.
type offDumpItem struct {
	scans int
	p     offProduct
}

// minHeap by scans keeps the Top most scanned products in O(Top) memory.
type offDumpHeap []offDumpItem

func (h offDumpHeap) Len() int            { return len(h) }
func (h offDumpHeap) Less(i, j int) bool  { return h[i].scans < h[j].scans }
func (h offDumpHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *offDumpHeap) Push(x interface{}) { *h = append(*h, x.(offDumpItem)) }
func (h *offDumpHeap) Pop() interface{} {
	old := *h
	it := old[len(old)-1]
	*h = old[:len(old)-1]
	return it
}

// ReadOFFDump streams the tab-separated export (gzip or plain) and returns the
// Top most scanned products sold in Country that declare carbohydrates,
// most scanned first.
func ReadOFFDump(r io.Reader, country string, top int) ([]offProduct, int, error) {
	br := bufio.NewReaderSize(r, 1<<20)
	if magic, _ := br.Peek(2); len(magic) == 2 && magic[0] == 0x1f && magic[1] == 0x8b {
		gz, err := gzip.NewReader(br)
		if err != nil {
			return nil, 0, err
		}
		defer func() { _ = gz.Close() }()
		br = bufio.NewReaderSize(gz, 1<<20)
	}
	sc := bufio.NewScanner(br)
	sc.Buffer(make([]byte, 1<<20), 64<<20)

	if !sc.Scan() {
		return nil, 0, fmt.Errorf("off dump: empty input")
	}
	col := map[string]int{}
	for i, name := range strings.Split(sc.Text(), "\t") {
		col[name] = i
	}
	for _, need := range []string{"code", "product_name", "countries_en", "unique_scans_n", "carbohydrates_100g"} {
		if _, ok := col[need]; !ok {
			return nil, 0, fmt.Errorf("off dump: missing column %q", need)
		}
	}
	get := func(f []string, name string) string {
		if i, ok := col[name]; ok && i < len(f) {
			return strings.TrimSpace(f[i])
		}
		return ""
	}
	num := func(f []string, name string) *float64 {
		v, err := strconv.ParseFloat(get(f, name), 64)
		if err != nil {
			return nil
		}
		return &v
	}

	country = strings.ToLower(country)
	h := &offDumpHeap{}
	matched := 0
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if !strings.Contains(strings.ToLower(get(f, "countries_en")), country) {
			continue
		}
		carb := num(f, "carbohydrates_100g")
		if carb == nil || get(f, "product_name") == "" || get(f, "code") == "" {
			continue
		}
		matched++
		scans, _ := strconv.Atoi(get(f, "unique_scans_n"))
		if top > 0 && h.Len() >= top && scans <= (*h)[0].scans {
			continue
		}
		brands, _ := json.Marshal(get(f, "brands"))
		it := offDumpItem{scans: scans, p: offProduct{
			Code:        get(f, "code"),
			ProductName: get(f, "product_name"),
			Categories:  get(f, "categories"),
			ImageURL:    get(f, "image_url"),
			Brands:      brands,
			Nutriments: &offNutriments{
				Energy100g: num(f, "energy-kcal_100g"),
				ProteinG:   num(f, "proteins_100g"),
				FatG:       num(f, "fat_100g"),
				CarbG:      carb,
				SugarG:     num(f, "sugars_100g"),
				FiberG:     num(f, "fiber_100g"),
				SaltG:      num(f, "salt_100g"),
			},
		}}
		heap.Push(h, it)
		if top > 0 && h.Len() > top {
			heap.Pop(h)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, matched, fmt.Errorf("off dump: %w", err)
	}
	items := []offDumpItem(*h)
	sort.Slice(items, func(i, j int) bool { return items[i].scans > items[j].scans })
	out := make([]offProduct, len(items))
	for i, it := range items {
		out[i] = it.p
	}
	return out, matched, nil
}

// ImportOFFDump imports the most scanned products of a country from the OFF
// CSV export read from r.
func ImportOFFDump(ctx context.Context, db *sql.DB, r io.Reader, opts OFFDumpOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "off"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	if opts.Country == "" {
		opts.Country = "Italy"
	}
	products, matched, err := ReadOFFDump(r, opts.Country, opts.Top)
	if err != nil {
		return nil, err
	}
	slog.Info("off dump parsed", "country", opts.Country, "with_carbs", matched, "selected", len(products))

	sourceID, err := lookupSourceID(ctx, db, opts.SourceCode)
	if err != nil {
		return nil, err
	}
	report := &ImportReport{}
	for i, p := range products {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		report.TotalRows++
		importOFFProduct(ctx, db, sourceID, p, opts.OFFOpts, report)
		if (i+1)%1000 == 0 {
			slog.Info("off dump progress", "done", i+1, "inserted", report.InsertedFoods,
				"updated", report.UpdatedFoods, "unchanged", report.Unchanged, "quarantined", report.Quarantined)
		}
	}
	return report, nil
}
