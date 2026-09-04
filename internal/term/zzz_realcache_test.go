package term

import (
	"fmt"
	"testing"
)

// Same question as TestZZZCacheKeyOmitsRawEndpoints, but against the ACTUAL
// shore palette drifting across a real day (hardcoded from
// internal/scape/palette.go's dayKeys, see the other zzz test for why it's
// duplicated here rather than imported).
func TestZZZRealPaletteCacheCollision(t *testing.T) {
	rampOnce.Do(rampInit)

	type key struct {
		t                         float64
		skyTop, skyHz, seaF, seaN RGB
	}
	keys := []key{
		{0.00, RGB{6, 8, 22}, RGB{40, 46, 78}, RGB{16, 24, 46}, RGB{30, 52, 84}},
		{0.25, RGB{0, 95, 135}, RGB{255, 175, 135}, RGB{0, 95, 95}, RGB{95, 95, 135}},
		{0.375, RGB{0, 95, 175}, RGB{175, 215, 255}, RGB{0, 95, 95}, RGB{135, 175, 255}},
		{0.50, RGB{0, 95, 175}, RGB{175, 215, 255}, RGB{0, 95, 95}, RGB{135, 175, 255}},
		{0.625, RGB{0, 95, 175}, RGB{255, 215, 175}, RGB{0, 95, 95}, RGB{135, 175, 255}},
		{0.75, RGB{0, 95, 175}, RGB{255, 135, 95}, RGB{0, 95, 95}, RGB{95, 135, 175}},
	}
	lerp := func(a, b RGB, f float64) RGB { return Lerp(a, b, f) }
	at := func(tt float64) (skyTop, skyHz, seaF, seaN RGB) {
		tt = tt - float64(int(tt))
		if tt < 0 {
			tt++
		}
		n := len(keys)
		for i := 0; i < n; i++ {
			a := keys[i]
			b := keys[(i+1)%n]
			hi := b.t
			if i == n-1 {
				hi = 1.0
			}
			if tt >= a.t && tt <= hi {
				span := hi - a.t
				if span <= 0 {
					return a.skyTop, a.skyHz, a.seaF, a.seaN
				}
				f := (tt - a.t) / span
				return lerp(a.skyTop, b.skyTop, f), lerp(a.skyHz, b.skyHz, f),
					lerp(a.seaF, b.seaF, f), lerp(a.seaN, b.seaN, f)
			}
		}
		return keys[0].skyTop, keys[0].skyHz, keys[0].seaF, keys[0].seaN
	}

	type bucket struct{ s, g int }
	type seenT struct {
		a, b RGB
		tt   float64
	}
	skySeen := map[bucket]seenT{}
	seaSeen := map[bucket]seenT{}
	skyDiff, seaDiff, skyColl, seaColl := 0, 0, 0, 0

	const samples = 24 * 60
	for i := 0; i < samples; i++ {
		tt := float64(i) / samples
		skyTop, skyHz, seaF, seaN := at(tt)

		sa, sg := rampEnd(skyTop), rampEnd(skyHz)
		if sa != sg {
			k := bucket{sa, sg}
			if prev, ok := skySeen[k]; ok {
				if prev.a != skyTop || prev.b != skyHz {
					skyColl++
					p1 := rampSearch(prev.a, prev.b, sa, sg)
					p2 := rampSearch(skyTop, skyHz, sa, sg)
					if !samePath(p1, p2) {
						skyDiff++
						if skyDiff <= 5 {
							fmt.Printf("SKY COLLISION bucket=%v first t=%.4f (%v->%v) vs t=%.4f (%v->%v)\n  p1=%s\n  p2=%s\n",
								k, prev.tt, prev.a, prev.b, tt, skyTop, skyHz, describe(p1), describe(p2))
						}
					}
				}
			} else {
				skySeen[k] = seenT{skyTop, skyHz, tt}
			}
		}

		qa, qg := rampEnd(seaF), rampEnd(seaN)
		if qa != qg {
			k := bucket{qa, qg}
			if prev, ok := seaSeen[k]; ok {
				if prev.a != seaF || prev.b != seaN {
					seaColl++
					p1 := rampSearch(prev.a, prev.b, qa, qg)
					p2 := rampSearch(seaF, seaN, qa, qg)
					if !samePath(p1, p2) {
						seaDiff++
					}
				}
			} else {
				seaSeen[k] = seenT{seaF, seaN, tt}
			}
		}
	}
	fmt.Println("sky bucket-collisions:", skyColl, "differing fresh path:", skyDiff)
	fmt.Println("sea bucket-collisions:", seaColl, "differing fresh path:", seaDiff)
}
