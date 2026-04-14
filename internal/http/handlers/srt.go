package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service/srt"
)

// SRT bundles the /admin/subject-rights/tickets endpoints.
type SRT struct {
	Svc      *srt.Service
	Notifier *srt.Notifier
	Erasure  srt.EraseInvoker // nullable in tests
}

// NewSRT constructs the handler.
func NewSRT(s *srt.Service, n *srt.Notifier, e srt.EraseInvoker) *SRT {
	return &SRT{Svc: s, Notifier: n, Erasure: e}
}

type createSRTReq struct {
	Kind           string `json:"kind"`
	RequesterEmail string `json:"requester_email"`
	Notes          string `json:"notes"`
}

// Create handles POST /admin/subject-rights/tickets.
func (h *SRT) Create(w http.ResponseWriter, r *http.Request) {
	var in createSRTReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	t, err := h.Svc.Create(r.Context(), srt.CreateInput{
		Kind: in.Kind, RequesterEmail: in.RequesterEmail, NotesMD: in.Notes,
	})
	if errors.Is(err, srt.ErrInvalidKind) {
		problem.Write(w, r, http.StatusBadRequest, "invalid_kind", "unknown ticket kind")
		return
	}
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	if h.Notifier != nil && in.RequesterEmail != "" {
		_ = h.Notifier.SendAck(r.Context(), t, in.RequesterEmail)
	}
	writeJSON(w, http.StatusCreated, t)
}

// List handles GET /admin/subject-rights/tickets.
func (h *SRT) List(w http.ResponseWriter, r *http.Request) {
	f := srt.ListFilter{
		Status:  r.URL.Query().Get("status"),
		Overdue: r.URL.Query().Get("overdue") == "true",
	}
	out, err := h.Svc.List(r.Context(), f)
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type patchSRTReq struct {
	Status       *string `json:"status"`
	AppendNotes  string  `json:"append_notes"`
	EvidencePath *string `json:"evidence_path"`
}

// Patch handles PATCH /admin/subject-rights/tickets/{id}.
func (h *SRT) Patch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var in patchSRTReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	t, err := h.Svc.Update(r.Context(), id, srt.PatchInput{
		Status: in.Status, AppendNotes: in.AppendNotes, EvidencePath: in.EvidencePath,
	})
	if errors.Is(err, srt.ErrNotFound) {
		problem.Write(w, r, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type closeSRTReq struct {
	Terminal       string `json:"terminal"`
	FinalNotes     string `json:"final_notes"`
	ErasureUserID  int64  `json:"erasure_user_id"`
}

// Close handles POST /admin/subject-rights/tickets/{id}/close.
func (h *SRT) Close(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}
	var in closeSRTReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	var (
		t   *srt.Ticket
		err error
	)
	if in.ErasureUserID > 0 && h.Erasure != nil {
		t, err = h.Svc.CloseErasure(r.Context(), id, in.ErasureUserID, h.Erasure)
	} else {
		t, err = h.Svc.Close(r.Context(), id, in.Terminal, in.FinalNotes)
	}
	if errors.Is(err, srt.ErrAlreadyClosed) {
		problem.Write(w, r, http.StatusConflict, "already_closed", "ticket already closed")
		return
	}
	if errors.Is(err, srt.ErrNotFound) {
		problem.Write(w, r, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
