package term

import (
	"fmt"
	"testing"
)

// Does the cache key omit the raw (a,b) that produced it, so a SECOND call
// landing on the same quantised (s,g) bucket -- but from a genuinely
// different true-colour line -- gets the FIRST call's path regardless of
// its own geometry? Search across many (s,g) pairs for two distinct raw
// lines that share a bucket and where a direct (uncached) rampSearch call
// would pick different paths for the two lines.
func TestZZZCacheKeyOmitsRawEndpoints(t *testing.T) {
	rampOnce.Do(rampInit)

	type sample struct {
		a, b RGB
	}
	// A grid of raw colour pairs, deliberately NOT snapped to cube entries,
	// so many distinct (a,b) share a quantised bucket the way a slowly
	// drifting palette does.
	var samples []sample
	for r := 0; r < 256; r += 23 {
		for g := 0; g < 256; g += 41 {
			for bl := 0; bl < 256; bl += 53 {
				a := RGB{uint8(r), uint8(g), uint8(bl)}
				b := RGB{uint8(255 - r), uint8((g + 90) % 256), uint8((bl + 60) % 256)}
				samples = append(samples, sample{a, b})
			}
		}
	}
	t.Logf("samples: %d", len(samples))

	type bucket struct{ s, g int }
	seen := map[bucket]sample{}
	diffs := 0
	checked := 0
	for _, s := range samples {
		sa, sg := rampEnd(s.a), rampEnd(s.b)
		if sa == sg {
			continue
		}
		key := bucket{sa, sg}
		first, ok := seen[key]
		if !ok {
			seen[key] = s
			continue
		}
		// Same bucket, different raw endpoints (or same -- skip identical).
		if first.a == s.a && first.b == s.b {
			continue
		}
		checked++
		p1 := rampSearch(first.a, first.b, sa, sg)
		p2 := rampSearch(s.a, s.b, sa, sg)
		if !samePath(p1, p2) {
			diffs++
			if diffs <= 5 {
				fmt.Printf("BUCKET s=%d g=%d\n  first a=%v b=%v -> %s\n  second a=%v b=%v -> %s\n",
					sa, sg, first.a, first.b, describe(p1), s.a, s.b, describe(p2))
			}
		}
	}
	fmt.Println("bucket collisions checked:", checked, " with a DIFFERENT fresh-search path:", diffs)
}

func samePath(a, b []RGB) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
