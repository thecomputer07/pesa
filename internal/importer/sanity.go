package importer

import (
	"fmt"
	"math"
)

// SanityCheck applies the two hard rules from 02-CONTEXT:
//
//  1. macro_sum = protein + fat + carb + water + ash + alcohol ≤ 105 g
//  2. |kcal_declared − (4p + 9f + 4c + 7alc)| / kcal_declared ≤ 0.20
//
// Missing critical values generate warn-level reasons without failing the
// check: carb_g being nil in particular is flagged because carbs drive the
// core calculator UX.
func SanityCheck(r FoodRow) SanityResult {
	res := SanityResult{OK: true, Severity: "warn"}

	// Warn: missing carbs (critical field for Pesa).
	if r.CarbG == nil {
		res.Reasons = append(res.Reasons, "missing_carb_g")
	}

	macroSum := 0.0
	add := func(p *float64) {
		if p != nil {
			macroSum += *p
		}
	}
	add(r.ProteinG)
	add(r.FatG)
	add(r.CarbG)
	add(r.WaterG)
	add(r.AshG)
	add(r.AlcoholG)

	if macroSum > 105.0 {
		res.OK = false
		res.Severity = "error"
		res.Reasons = append(res.Reasons, fmt.Sprintf("macro_sum_overflow: %.2f", macroSum))
	}

	if r.Kcal != nil && *r.Kcal > 0 {
		computed := 0.0
		if r.ProteinG != nil {
			computed += 4 * *r.ProteinG
		}
		if r.FatG != nil {
			computed += 9 * *r.FatG
		}
		if r.CarbG != nil {
			computed += 4 * *r.CarbG
		}
		if r.AlcoholG != nil {
			computed += 7 * *r.AlcoholG
		}
		drift := math.Abs(*r.Kcal-computed) / *r.Kcal
		if drift > 0.20 {
			res.OK = false
			res.Severity = "error"
			res.Reasons = append(res.Reasons,
				fmt.Sprintf("kcal_mismatch: declared=%.1f computed=%.1f drift=%.0f%%",
					*r.Kcal, computed, drift*100))
		}
	}

	return res
}
