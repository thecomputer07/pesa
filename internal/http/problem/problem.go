// Package problem implements RFC 7807 problem+json responses.
package problem

import (
	"encoding/json"
	"net/http"
)

// Problem is the RFC 7807 envelope.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// Write emits a problem+json response with the given code and detail.
func Write(w http.ResponseWriter, r *http.Request, status int, code, detail string) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Problem{
		Type:     "https://pesadia.com/problems/" + code,
		Title:    code,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.Path,
	})
}
