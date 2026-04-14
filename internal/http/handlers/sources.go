package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/repo"
)

// Sources serves GET /api/v1/sources.
type Sources struct {
	DB *sql.DB
}

// NewSources builds the handler bound to the given DB.
func NewSources(db *sql.DB) *Sources { return &Sources{DB: db} }

type sourceDTO struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	License     string  `json:"license"`
	LicenseURL  *string `json:"license_url,omitempty"`
	Attribution string  `json:"attribution"`
	SourceURL   *string `json:"source_url,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	Contact     *string `json:"contact,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type sourcesResponseDTO struct {
	Sources []sourceDTO `json:"sources"`
}

// List returns every food_sources row as JSON with a 1h cache.
func (h *Sources) List(w http.ResponseWriter, r *http.Request) {
	rows, err := repo.ListFoodSources(r.Context(), h.DB)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "sources list failed")
		return
	}
	out := sourcesResponseDTO{Sources: make([]sourceDTO, 0, len(rows))}
	for _, s := range rows {
		out.Sources = append(out.Sources, sourceDTO{
			Code:        s.Code,
			Name:        s.Name,
			License:     s.License,
			LicenseURL:  nullStrPtr(s.LicenseURL),
			Attribution: s.AttributionText,
			SourceURL:   nullStrPtr(s.SourceURL),
			Scope:       nullStrPtr(s.Scope),
			Contact:     nullStrPtr(s.Contact),
			Notes:       nullStrPtr(s.Notes),
		})
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}
