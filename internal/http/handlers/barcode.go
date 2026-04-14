package handlers

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service/off"
)

// BarcodeLookup is the narrow interface the handler depends on; it allows
// tests to substitute a fake without a real HTTP/cache stack.
type BarcodeLookup interface {
	LookupByEAN(ctx context.Context, ean string) (off.FoodDTO, error)
}

// Barcode exposes GET /api/v1/foods/barcode/{ean}.
type Barcode struct {
	Svc BarcodeLookup
}

// NewBarcode builds a Barcode handler bound to the given service.
func NewBarcode(svc BarcodeLookup) *Barcode { return &Barcode{Svc: svc} }

var eanRegex = regexp.MustCompile(`^\d{8,14}$`)

// Lookup validates the EAN, zero-pads to 13, and delegates to the service.
func (h *Barcode) Lookup(w http.ResponseWriter, r *http.Request) {
	raw := strings.TrimSpace(chi.URLParam(r, "ean"))
	if !eanRegex.MatchString(raw) {
		problem.Write(w, r, http.StatusBadRequest, "invalid-ean", "ean must be 8–14 digits")
		return
	}
	ean := zeroPadTo13(raw)
	dto, err := h.Svc.LookupByEAN(r.Context(), ean)
	if err != nil {
		if errors.Is(err, off.ErrNotFound) {
			problem.Write(w, r, http.StatusNotFound, "barcode-not-found", "no OFF product for ean")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "barcode lookup failed")
		return
	}
	writeJSONCached(w, http.StatusOK, dto)
}

func zeroPadTo13(s string) string {
	if len(s) >= 13 {
		// If it's 14 digits starting with "0", drop the leading zero (GTIN-14).
		if len(s) == 14 && s[0] == '0' {
			return s[1:]
		}
		return s
	}
	return strings.Repeat("0", 13-len(s)) + s
}
