// Package i18n wires the nicksnyder/go-i18n bundle with IT + EN catalogs.
package i18n

import (
	"embed"
	"encoding/json"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localesFS embed.FS

// NewBundle loads all catalogs at startup. IT is the default.
func NewBundle() (*goi18n.Bundle, error) {
	b := goi18n.NewBundle(language.Italian)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := localesFS.ReadFile("locales/" + e.Name())
		if err != nil {
			return nil, err
		}
		if _, err := b.ParseMessageFileBytes(data, e.Name()); err != nil {
			return nil, err
		}
	}
	return b, nil
}
