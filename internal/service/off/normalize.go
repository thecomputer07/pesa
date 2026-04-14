package off

import (
	"encoding/json"
	"fmt"
)

// FoodDTO mirrors the foodDetailDTO shape used by /api/v1/foods handlers.
// ID=0 signals a virtual (non-persisted) OFF food.
type FoodDTO struct {
	ID          int64      `json:"id"`
	NameIT      string     `json:"name_it"`
	NameEN      *string    `json:"name_en,omitempty"`
	Category    *string    `json:"category,omitempty"`
	State       string     `json:"state"`
	EdiblePct   float64    `json:"edible_pct"`
	Kcal        *float64   `json:"kcal,omitempty"`
	ProteinG    *float64   `json:"protein_g,omitempty"`
	FatG        *float64   `json:"fat_g,omitempty"`
	CarbG       *float64   `json:"carb_g,omitempty"`
	SugarG      *float64   `json:"sugar_g,omitempty"`
	FiberG      *float64   `json:"fiber_g,omitempty"`
	Attribution string     `json:"attribution"`
	License     string     `json:"license"`
	LicenseURL  *string    `json:"license_url,omitempty"`
	SourceURL   *string    `json:"source_url,omitempty"`
	Source      SourceDTO  `json:"source"`
	Images      []ImageDTO `json:"images"`
}

// SourceDTO mirrors the shape foods.Search emits so clients can decode both.
type SourceDTO struct {
	Code       string  `json:"code"`
	License    string  `json:"license"`
	LicenseURL *string `json:"license_url,omitempty"`
}

// ImageDTO is a minimal image entry; License/Attribution are filled when the
// image is served via the local /static/images cache (plan-02).
type ImageDTO struct {
	URL         string  `json:"url"`
	License     *string `json:"license,omitempty"`
	Attribution *string `json:"attribution,omitempty"`
	IsExternal  bool    `json:"is_external"`
}

// offRaw captures only the fields we consume from OFF's v2 product endpoint.
type offRaw struct {
	Status  int        `json:"status"`   // 1 = found, 0 = not_found
	Code    string     `json:"code"`
	Product *offProd   `json:"product"`
}

type offProd struct {
	ProductName   string      `json:"product_name"`
	ProductNameIT string      `json:"product_name_it"`
	ProductNameEN string      `json:"product_name_en"`
	Brands        string      `json:"brands"`
	Quantity      string      `json:"quantity"`
	Categories    string      `json:"categories"`
	ImageFrontURL string      `json:"image_front_small_url"`
	Nutriments    offNutris   `json:"nutriments"`
}

type offNutris struct {
	Energy   *float64 `json:"energy-kcal_100g"`
	Protein  *float64 `json:"proteins_100g"`
	Fat      *float64 `json:"fat_100g"`
	Carb     *float64 `json:"carbohydrates_100g"`
	Sugar    *float64 `json:"sugars_100g"`
	Fiber    *float64 `json:"fiber_100g"`
}

// Constants for OFF licensing.
const (
	offLicense        = "ODbL-1.0"
	offLicenseURL     = "https://opendatacommons.org/licenses/odbl/1-0/"
	offAttribution    = "Open Food Facts contributors — ODbL"
	offProductURLTmpl = "https://world.openfoodfacts.org/product/%s"
)

// FromOFF maps a raw OFF product JSON payload onto a FoodDTO.
// Returns ErrNotFound when the payload has no product node.
func FromOFF(raw []byte, ean string) (FoodDTO, error) {
	var r offRaw
	if err := json.Unmarshal(raw, &r); err != nil {
		return FoodDTO{}, fmt.Errorf("off: decode: %w", err)
	}
	if r.Product == nil || r.Status == 0 {
		return FoodDTO{}, ErrNotFound
	}
	p := r.Product

	name := firstNonEmpty(p.ProductNameIT, p.ProductNameEN, p.ProductName)
	if name == "" {
		name = ean // at worst surface the EAN so UI has something to render
	}

	var nameEN *string
	if p.ProductNameEN != "" {
		v := p.ProductNameEN
		nameEN = &v
	}

	category := firstNonEmpty(p.Brands, p.Categories)
	var catPtr *string
	if category != "" {
		v := category
		catPtr = &v
	}

	licURL := offLicenseURL
	srcURL := fmt.Sprintf(offProductURLTmpl, ean)

	dto := FoodDTO{
		ID:          0,
		NameIT:      name,
		NameEN:      nameEN,
		Category:    catPtr,
		State:       "packaged",
		EdiblePct:   100,
		Kcal:        p.Nutriments.Energy,
		ProteinG:    p.Nutriments.Protein,
		FatG:        p.Nutriments.Fat,
		CarbG:       p.Nutriments.Carb,
		SugarG:      p.Nutriments.Sugar,
		FiberG:      p.Nutriments.Fiber,
		Attribution: offAttribution,
		License:     offLicense,
		LicenseURL:  &licURL,
		SourceURL:   &srcURL,
		Source: SourceDTO{
			Code:       "off",
			License:    offLicense,
			LicenseURL: &licURL,
		},
		Images: []ImageDTO{},
	}
	if p.ImageFrontURL != "" {
		attr := offAttribution
		lic := offLicense
		dto.Images = append(dto.Images, ImageDTO{
			URL:         p.ImageFrontURL,
			License:     &lic,
			Attribution: &attr,
			IsExternal:  true,
		})
	}
	return dto, nil
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
