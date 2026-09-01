package scape

import "github.com/donlucasx/xscapes/internal/term"

// Palette is every colour the shore needs at one moment of the day. Keeping
// them together means a new time of day is one struct, not thirty edits.
type Palette struct {
	SkyTop, SkyHorizon term.RGB
	SeaFar, SeaNear    term.RGB
	Foam               term.RGB
	WetSand, SandNear  term.RGB
	Grain              term.RGB
	Star, Moon         term.RGB
	Glitter            term.RGB

	// How present the night furniture is. Stars and the moon do not vanish at
	// noon so much as wash out, and a hard cut reads as a bug.
	StarVis, MoonVis float64
}

// dayKeys are keyframes around the clock; everything between is interpolated.
var dayKeys = []struct {
	t float64
	p Palette
}{
	{0.00, Palette{ // midnight
		SkyTop: term.RGB{R: 6, G: 8, B: 22}, SkyHorizon: term.RGB{R: 40, G: 46, B: 78},
		SeaFar: term.RGB{R: 16, G: 24, B: 46}, SeaNear: term.RGB{R: 30, G: 52, B: 84},
		Foam:    term.RGB{R: 214, G: 232, B: 238},
		WetSand: term.RGB{R: 58, G: 51, B: 45}, SandNear: term.RGB{R: 76, G: 65, B: 54},
		Grain: term.RGB{R: 116, G: 101, B: 84},
		Star:  term.RGB{R: 206, G: 216, B: 240}, Moon: term.RGB{R: 242, G: 238, B: 214},
		Glitter: term.RGB{R: 236, G: 232, B: 200},
		StarVis: 1.0, MoonVis: 1.0,
	}},
	{0.25, Palette{ // dawn
		SkyTop: term.RGB{R: 44, G: 52, B: 96}, SkyHorizon: term.RGB{R: 232, G: 152, B: 122},
		SeaFar: term.RGB{R: 34, G: 44, B: 72}, SeaNear: term.RGB{R: 92, G: 92, B: 116},
		Foam:    term.RGB{R: 250, G: 226, B: 210},
		WetSand: term.RGB{R: 92, G: 76, B: 68}, SandNear: term.RGB{R: 152, G: 126, B: 102},
		Grain: term.RGB{R: 190, G: 162, B: 134},
		Star:  term.RGB{R: 210, G: 218, B: 240}, Moon: term.RGB{R: 244, G: 238, B: 226},
		Glitter: term.RGB{R: 255, G: 214, B: 180},
		StarVis: 0.25, MoonVis: 0.45,
	}},
	{0.50, Palette{ // noon
		SkyTop: term.RGB{R: 54, G: 116, B: 196}, SkyHorizon: term.RGB{R: 164, G: 204, B: 238},
		SeaFar: term.RGB{R: 24, G: 74, B: 116}, SeaNear: term.RGB{R: 58, G: 134, B: 166},
		Foam:    term.RGB{R: 244, G: 252, B: 255},
		WetSand: term.RGB{R: 124, G: 104, B: 86}, SandNear: term.RGB{R: 202, G: 176, B: 142},
		Grain: term.RGB{R: 232, G: 208, B: 174},
		Star:  term.RGB{R: 220, G: 228, B: 244}, Moon: term.RGB{R: 240, G: 240, B: 240},
		Glitter: term.RGB{R: 255, G: 255, B: 236},
		StarVis: 0.0, MoonVis: 0.10,
	}},
	{0.75, Palette{ // dusk
		SkyTop: term.RGB{R: 46, G: 38, B: 88}, SkyHorizon: term.RGB{R: 244, G: 138, B: 88},
		SeaFar: term.RGB{R: 30, G: 32, B: 64}, SeaNear: term.RGB{R: 100, G: 74, B: 98},
		Foam:    term.RGB{R: 252, G: 208, B: 178},
		WetSand: term.RGB{R: 82, G: 64, B: 56}, SandNear: term.RGB{R: 140, G: 110, B: 86},
		Grain: term.RGB{R: 178, G: 144, B: 112},
		Star:  term.RGB{R: 214, G: 220, B: 242}, Moon: term.RGB{R: 250, G: 232, B: 206},
		Glitter: term.RGB{R: 255, G: 192, B: 132},
		StarVis: 0.35, MoonVis: 0.55,
	}},
}

func lerpPalette(a, b Palette, t float64) Palette {
	l := func(x, y term.RGB) term.RGB { return term.Lerp(x, y, t) }
	f := func(x, y float64) float64 { return x + (y-x)*t }
	return Palette{
		SkyTop: l(a.SkyTop, b.SkyTop), SkyHorizon: l(a.SkyHorizon, b.SkyHorizon),
		SeaFar: l(a.SeaFar, b.SeaFar), SeaNear: l(a.SeaNear, b.SeaNear),
		Foam:    l(a.Foam, b.Foam),
		WetSand: l(a.WetSand, b.WetSand), SandNear: l(a.SandNear, b.SandNear),
		Grain: l(a.Grain, b.Grain),
		Star:  l(a.Star, b.Star), Moon: l(a.Moon, b.Moon),
		Glitter: l(a.Glitter, b.Glitter),
		StarVis: f(a.StarVis, b.StarVis), MoonVis: f(a.MoonVis, b.MoonVis),
	}
}

// PaletteAt wraps around the clock, so 0.99 blends back into midnight rather
// than clamping to dusk.
func PaletteAt(t float64) Palette {
	t = t - float64(int(t))
	if t < 0 {
		t++
	}
	n := len(dayKeys)
	for i := 0; i < n; i++ {
		a := dayKeys[i]
		b := dayKeys[(i+1)%n]
		hi := b.t
		if i == n-1 {
			hi = 1.0
		}
		if t >= a.t && t <= hi {
			span := hi - a.t
			if span <= 0 {
				return a.p
			}
			return lerpPalette(a.p, b.p, (t-a.t)/span)
		}
	}
	return dayKeys[0].p
}
