// Package legal exposes compile-time legal constants.
package legal

// PrivacyPolicyVersion is the version string stored in users.privacy_consent_version
// at signup. Bumped whenever .planning/legal/privacy-policy.it.md ships a material change.
const PrivacyPolicyVersion = "1.0"

// CookiePolicyVersion mirrors cookie-policy.it.md versioning.
const CookiePolicyVersion = "1.0"
