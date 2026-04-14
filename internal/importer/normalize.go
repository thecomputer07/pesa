package importer

// Normalize converts row to per-100g-edible-raw basis.
//
// CREA publishes per-100g-edible already, so BasisEdible is a passthrough.
// When basis is BasisAsPurchased, nutrient mass values are scaled up by
// 100/EdiblePct so the resulting numbers describe 100g of *edible* portion.
func Normalize(row FoodRow, basis Basis) FoodRow {
	if basis == BasisEdible || row.EdiblePct <= 0 || row.EdiblePct >= 100.0 {
		if row.EdiblePct <= 0 {
			row.EdiblePct = 100.0
		}
		return row
	}
	factor := 100.0 / row.EdiblePct
	scale := func(p *float64) *float64 {
		if p == nil {
			return nil
		}
		v := *p * factor
		return &v
	}
	row.Kcal = scale(row.Kcal)
	row.ProteinG = scale(row.ProteinG)
	row.FatG = scale(row.FatG)
	row.CarbG = scale(row.CarbG)
	row.SugarG = scale(row.SugarG)
	row.FiberG = scale(row.FiberG)
	row.WaterG = scale(row.WaterG)
	row.AshG = scale(row.AshG)
	row.AlcoholG = scale(row.AlcoholG)
	row.SaltG = scale(row.SaltG)
	return row
}
