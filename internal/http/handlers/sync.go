package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service/sync"
)

// Sync groups /api/v1/sync endpoints.
type Sync struct {
	svc *sync.Service
}

// NewSync constructs the handler.
func NewSync(svc *sync.Service) *Sync { return &Sync{svc: svc} }

const syncMaxBody = 1 << 20 // 1 MiB (SEC-22 / PERF-26)

// Post handles POST /api/v1/sync — batch upsert.
func (h *Sync) Post(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, syncMaxBody)

	// Pre-decode to a map so we can reject the "foods" key explicitly (server-wins catalog).
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		var mbErr *http.MaxBytesError
		if errors.As(err, &mbErr) || strings.Contains(err.Error(), "http: request body too large") {
			problem.Write(w, r, http.StatusRequestEntityTooLarge, "body-too-large", "sync body exceeds 1MB limit")
			return
		}
		problem.Write(w, r, http.StatusBadRequest, "invalid-json", err.Error())
		return
	}
	if _, hasFoods := raw["foods"]; hasFoods {
		problem.Write(w, r, http.StatusUnprocessableEntity, "foods-readonly",
			"foods are server-managed; do not push them via /api/v1/sync")
		return
	}

	// Re-marshal the accepted keys into BatchPayload.
	var p sync.BatchPayload
	if len(raw) > 0 {
		filtered, _ := json.Marshal(raw)
		if err := json.Unmarshal(filtered, &p); err != nil {
			problem.Write(w, r, http.StatusBadRequest, "invalid-shape", err.Error())
			return
		}
	}

	// TODO(validate): run go-playground/validator on p per SEC-25.
	// Deferred to avoid introducing the dep transitively here; basic type-level
	// guards (required fields, positive grams, enum slot) are enforced by the DB
	// CHECK constraints + NOT NULL columns in plan-01 schema.

	resp, err := h.svc.UpsertBatch(r.Context(), userID, p)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "sync-failed", "apply failed")
		return
	}
	writeJSONNoCache(w, http.StatusOK, resp)
}

// Get handles GET /api/v1/sync?since=<ms> — delta pull.
func (h *Sync) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	if userID <= 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
	if since < 0 {
		since = 0
	}
	resp, err := h.svc.Pull(r.Context(), userID, since)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "pull-failed", "delta pull failed")
		return
	}
	writeJSONNoCache(w, http.StatusOK, resp)
}

func writeJSONNoCache(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
