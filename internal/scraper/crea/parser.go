package crea

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ParseDetailHTML extracts a CREARecord from a CREA detail page's HTML body.
// It targets tables with class "tabella-composizione" where each row has
// <td class="label"> and <td class="value">.
func ParseDetailHTML(html []byte, sourceURL string) (*CREARecord, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("goquery: %w", err)
	}

	rec := &CREARecord{
		URL:         sourceURL,
		ExtraFields: map[string]string{},
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	sum := sha256.Sum256(html)
	rec.ContentHash = hex.EncodeToString(sum[:])

	// External ID = last non-empty path segment.
	if u, uErr := url.Parse(sourceURL); uErr == nil {
		rec.ExternalID = strings.Trim(path.Base(u.Path), "/")
	}

	// Name: the <h1 class="alimento-nome"> or <h1> fallback.
	if h := strings.TrimSpace(doc.Find("h1.alimento-nome").First().Text()); h != "" {
		rec.NameIT = h
	} else {
		rec.NameIT = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Category: <span class="categoria">.
	rec.Category = strings.TrimSpace(doc.Find("span.categoria").First().Text())

	// Walk composition tables.
	doc.Find("table.tabella-composizione tr").Each(func(_ int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("td.label").Text())
		value := strings.TrimSpace(s.Find("td.value").Text())
		if label == "" {
			return
		}
		assignField(rec, label, value)
	})

	if rec.NameIT == "" {
		return nil, fmt.Errorf("parse: missing name")
	}
	return rec, nil
}

// ParseIndexHTML extracts absolute detail URLs from an index page.
func ParseIndexHTML(html []byte, baseURL string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	out := []string{}
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		if !strings.Contains(href, "/valori-nutrizionali/") {
			return
		}
		ref, err := url.Parse(href)
		if err != nil {
			return
		}
		abs := base.ResolveReference(ref).String()
		if _, dup := seen[abs]; dup {
			return
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
	})
	return out, nil
}

var numRe = regexp.MustCompile(`-?\d+(?:[.,]\d+)?`)

// parseItalianNumber parses "12,34" or "12.34" to float64. Returns false if no digits.
func parseItalianNumber(s string) (float64, bool) {
	m := numRe.FindString(s)
	if m == "" {
		return 0, false
	}
	m = strings.ReplaceAll(m, ",", ".")
	f, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func assignField(rec *CREARecord, label, value string) {
	normLabel := strings.ToLower(label)
	normLabel = strings.TrimSuffix(normLabel, ":")
	normLabel = strings.TrimSpace(normLabel)

	f, ok := parseItalianNumber(value)
	ptr := func(v float64) *float64 { return &v }

	switch {
	case strings.HasPrefix(normLabel, "parte edibile"):
		if ok {
			rec.EdiblePct = ptr(f)
		}
	case strings.Contains(normLabel, "energia") && strings.Contains(normLabel, "kcal"):
		if ok {
			rec.KcalPer100g = ptr(f)
		}
	case strings.HasPrefix(normLabel, "proteine"):
		if ok {
			rec.ProteinG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "lipidi") || strings.HasPrefix(normLabel, "grassi"):
		if ok {
			rec.FatG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "carboidrati"):
		if ok {
			rec.CarbG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "zuccheri"):
		if ok {
			rec.SugarG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "fibra"):
		if ok {
			rec.FiberG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "acqua"):
		if ok {
			rec.WaterG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "ceneri"):
		if ok {
			rec.AshG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "alcol") || strings.HasPrefix(normLabel, "alcool"):
		if ok {
			rec.AlcoholG = ptr(f)
		}
	case strings.HasPrefix(normLabel, "sale") || strings.HasPrefix(normLabel, "sodio"):
		if ok {
			rec.SaltG = ptr(f)
		}
	default:
		rec.ExtraFields[label] = value
	}
}
