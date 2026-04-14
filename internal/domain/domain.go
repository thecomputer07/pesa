// Package domain holds core domain types (users, meals, recipes).
// Phase 1 introduces only the user / auth subset.
package domain

import "time"

// User mirrors the users table. Password hash and tokens remain server-side.
type User struct {
	ID                     int64
	Email                  string
	EmailVerifiedAt        *time.Time
	Locale                 string
	PrivacyConsentAt       time.Time
	PrivacyConsentVersion  string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Status                 string // "active" | "restricted" | "deleted"
}
