package companion

// nearFillHidden removes the features a body hides at 12x7 and would reveal at
// 24x14, and leaves the ones it should reveal alone.
//
// THE PROBLEM. ToQuadrant ORs source rows 2k and 2k+1 before packing, so within
// a pair only the union reaches the screen. Doubling the source for the nearest
// rung puts each row of the pair on its own half-row, where both become
// visible. Most of what surfaces is wanted -- every tapered edge on this animal
// is a pair whose two rows differ by a column or two, and at 2x that taper is
// exactly the extra detail the bigger sprite is for.
//
// One thing that surfaces is NOT wanted, and he caught it: CatBody row 11 is a
// muzzle gap that row 10 covers completely, so the cat has never had a mouth on
// screen -- and a mouth appeared at 24x14. All four cat bodies carry it.
//
// THE RULE: fill an INTERIOR run of empty cells that the pair's other row
// covers completely -- but ONLY on the muzzle pair.
//
// ⚠ IT IS TARGETED, AND THAT IS A FINDING RATHER THAN LAZINESS. Two blunter
// rules were written and both were measured wrong on the render:
//
//   - "Fill wherever the partner has ink" ate two columns out of the EAR GAP.
//   - "Fill every covered interior run, on every pair" ate the NOTCH BETWEEN
//     THE EARS as well as the mouth. The notch and the mouth are structurally
//     identical -- an interior gap that the partner row covers -- so no
//     general rule can keep one and drop the other. The notch is a real
//     feature of a cat's head and is wanted at 24x14; the mouth is not.
//
// So the row band is named. muzzlePair is the pair immediately under the eyes
// (rows 8..9), which is the only place on this animal a mouth can be.
const muzzlePair = 10

func nearFillHidden(rows []string) []string {
	out := make([]string, len(rows))
	copy(out, rows)
	if muzzlePair+1 < len(rows) {
		out[muzzlePair] = fillCovered(rows[muzzlePair], rows[muzzlePair+1])
		out[muzzlePair+1] = fillCovered(rows[muzzlePair+1], rows[muzzlePair])
	}
	return out
}

// fillCovered fills a's interior empty runs that b covers completely.
func fillCovered(a, b string) string {
	ra, rb := []rune(a), []rune(b)
	if len(ra) != len(rb) {
		return a
	}
	for x := 0; x < len(ra); x++ {
		if ra[x] == '#' {
			continue
		}
		// The run of empties starting here.
		end := x
		for end < len(ra) && ra[end] != '#' {
			end++
		}
		// Interior means ink on both sides within this row.
		interior := x > 0 && ra[x-1] == '#' && end < len(ra) && ra[end] == '#'
		if interior {
			covered := true
			for k := x; k < end; k++ {
				if rb[k] != '#' {
					covered = false
					break
				}
			}
			if covered {
				for k := x; k < end; k++ {
					ra[k] = '#'
				}
			}
		}
		x = end
	}
	return string(ra)
}
