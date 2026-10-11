package importer

import (
	"fmt"
	"math"
)

// SanityCheck applies the two hard rules from 02-CONTEXT:
//
//  1. macro_sum = protein + fat + carb + water + ash + alcohol ≤ 112 g
//
// Over 100 is normal: CREA gives available carbohydrate as monosaccharide
// equivalents (starch x 1.1), so potato starch sums to 108 g. The ceiling
// only has to catch real errors (a value read ten times too large).
//  2. |kcal_declared − (4p + 9f + 4c + 2fibre + 7alc)| / kcal_declared ≤ 0.20
//
// Fibre counts 2 kcal/g (Reg. UE 1169/2011, all. XIV): without it every
// low-calorie vegetable (artichokes, fennel, lemons…) failed the check.
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

	if macroSum > 112.0 {
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
		if r.FiberG != nil {
			computed += 2 * *r.FiberG
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
