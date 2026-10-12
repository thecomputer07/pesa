package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/repo"
)

// Catalog serves the public food catalog for offline use (PWA / app):
// GET /api/v1/catalog downloads everything, ?since=<cursor> only what changed.
// The body is the same for every user, so it is cacheable by browsers and
// Cloudflare; the ETag is the catalog state, so an unchanged catalog costs a 304.
type Catalog struct {
	DB *sql.DB

	mu       sync.Mutex
	fullAt   repo.CatalogState // state the cached full body was built for
	fullBody []byte
}

// NewCatalog builds a Catalog handler.
func NewCatalog(db *sql.DB) *Catalog { return &Catalog{DB: db} }

type catalogDTO struct {
	Cursor  int64                `json:"cursor"`
	Count   int                  `json:"count"`
	Full    bool                 `json:"full"`
	Columns []string             `json:"columns"`
	Rows    []repo.CatalogRow    `json:"rows"`
	Gone    []int64              `json:"gone"`
	Sources []repo.CatalogSource `json:"sources"`
}

// Get handles GET /api/v1/catalog?since=<ms>.
func (h *Catalog) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	if since < 0 {
		since = 0
	}
	st, err := repo.LoadCatalogState(ctx, h.DB)
	if err != nil {
		slog.Error("catalog state", "err", err)
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "catalog unavailable")
		return
	}

	etag := fmt.Sprintf(`"cat-%d-%d-%d"`, st.CursorMs, st.Count, since)
	w.Header().Set("ETag", etag)
	// Browsers revalidate (cheap 304); Cloudflare keeps it for a while.
	w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate, s-maxage=600")
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var body []byte
	if since == 0 {
		body, err = h.full(r, st)
	} else {
		body, err = h.build(r, st, since)
	}
	if err != nil {
		slog.Error("catalog", "since", since, "err", err)
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "catalog unavailable")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

// full returns the whole catalog, rebuilt only when the catalog state changes.
func (h *Catalog) full(r *http.Request, st repo.CatalogState) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.fullBody != nil && h.fullAt == st {
		return h.fullBody, nil
	}
	body, err := h.build(r, st, 0)
	if err != nil {
		return nil, err
	}
	h.fullAt, h.fullBody = st, body
	return body, nil
}

func (h *Catalog) build(r *http.Request, st repo.CatalogState, since int64) ([]byte, error) {
	out := catalogDTO{
		Cursor: st.CursorMs, Count: st.Count, Full: since == 0,
		Columns: repo.CatalogColumns, Rows: []repo.CatalogRow{}, Gone: []int64{},
	}
	if since == 0 || since < st.CursorMs {
		rows, gone, err := repo.LoadCatalog(r.Context(), h.DB, since)
		if err != nil {
			return nil, err
		}
		out.Rows, out.Gone = rows, gone
	}
	sources, err := repo.LoadCatalogSources(r.Context(), h.DB)
	if err != nil {
		return nil, err
	}
	out.Sources = sources
	return json.Marshal(out)
}
