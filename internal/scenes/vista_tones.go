package scenes

import (
	"math"

	"github.com/donlucasx/xscapes/internal/scape"
	"github.com/donlucasx/xscapes/internal/term"
)

// THE VISTA'S DAYTIME TONES (s36, 2026-09-16). His note: "can we pick more
// pleasant tones for the daytime cycle". The day sky the vista inherits is
// the shore's (skyDeep, an electric zenith, over skyPale) and its meadow is
// an olive the cube holds at full saturation. A tone replaces the day sky's
// zenith and horizon, the meadow's two greens and the scrub, as one set,
// every colour a cube entry so it survives quantisation as drawn. It is
// applied by DAY only, faded in from dawn and out to dusk, so the peach and
// orange horizons and the night are untouched. Tone 0 is today, and the
// study clip's guard holds on it.
//
// ⚠ The afternoon. The shore's mid-afternoon horizon is skyWarm (255,215,
// 175), and the ramp from any blue zenith down to a warm pale runs through
// grey: at 14:00 today's sky is a grey haze band over the ranges, and the
// first cut of the cool tones inherited it. The cool tones keep a cool
// horizon until the dusk keyframe takes over; only "warm" means it.
type VistaTone struct {
	Name, Note string
	SkyTop     term.RGB
	// The horizon at mid-morning, noon and mid-afternoon (the shore's
	// keyframes at 0.375, 0.5 and 0.625), blended between.
	HzMorning, HzNoon, HzAfternoon term.RGB
	MeadowTop, MeadowBot, Scrub    term.RGB
	// WarmMid, when set, is a third stop the sky passes through at the
	// hours when the horizon is warm (dawn and dusk), so the ramp from a
	// blue zenith to a peach or orange horizon does not cross grey: that
	// crossing is the band of grey rows over the ranges at 17:30 in every
	// tone, today's included (his crop, 2026-09-16 12:31).
	WarmMid term.RGB
}

var VistaTones = []VistaTone{
	{Name: "today", Note: "the shore's day sky: an electric zenith over a pale horizon; olive meadow, acid scrub",
		SkyTop: term.RGB{R: 0, G: 95, B: 175}, HzMorning: term.RGB{R: 175, G: 215, B: 255}, HzNoon: term.RGB{R: 175, G: 215, B: 255}, HzAfternoon: term.RGB{R: 255, G: 215, B: 175},
		MeadowTop: term.RGB{R: 95, G: 135, B: 0}, MeadowBot: term.RGB{R: 95, G: 95, B: 0}, Scrub: term.RGB{R: 135, G: 175, B: 0}},
	{Name: "soft alpine", Note: "a softer, lighter blue at the zenith, a pale horizon that stays cool until dusk; a sage meadow over dark green, soft green scrub",
		SkyTop: term.RGB{R: 95, G: 135, B: 215}, HzMorning: term.RGB{R: 175, G: 215, B: 255}, HzNoon: term.RGB{R: 175, G: 215, B: 255}, HzAfternoon: term.RGB{R: 175, G: 215, B: 215},
		MeadowTop: term.RGB{R: 95, G: 135, B: 95}, MeadowBot: term.RGB{R: 0, G: 95, B: 0}, Scrub: term.RGB{R: 135, G: 175, B: 95}},
	{Name: "pastel", Note: "a milky sky, light blue to near-white lavender all day; a light sage meadow, pale scrub",
		SkyTop: term.RGB{R: 135, G: 175, B: 215}, HzMorning: term.RGB{R: 215, G: 215, B: 255}, HzNoon: term.RGB{R: 215, G: 215, B: 255}, HzAfternoon: term.RGB{R: 215, G: 215, B: 255},
		MeadowTop: term.RGB{R: 135, G: 175, B: 95}, MeadowBot: term.RGB{R: 95, G: 135, B: 95}, Scrub: term.RGB{R: 175, G: 215, B: 135}},
	{Name: "warm", Note: "a steel-blue zenith over a warm pale horizon all day; a gold-olive meadow, straw scrub -- late summer",
		SkyTop: term.RGB{R: 95, G: 135, B: 175}, HzMorning: term.RGB{R: 175, G: 215, B: 215}, HzNoon: term.RGB{R: 215, G: 215, B: 175}, HzAfternoon: term.RGB{R: 215, G: 215, B: 175},
		MeadowTop: term.RGB{R: 135, G: 135, B: 0}, MeadowBot: term.RGB{R: 95, G: 95, B: 0}, Scrub: term.RGB{R: 175, G: 175, B: 95}},
	{Name: "deep and cool", Note: "a deep blue zenith over a mid-blue horizon that stays cool until dusk, the mountain sky; the olive meadow over dark green, today's scrub",
		SkyTop: term.RGB{R: 0, G: 95, B: 135}, HzMorning: term.RGB{R: 135, G: 175, B: 215}, HzNoon: term.RGB{R: 175, G: 215, B: 215}, HzAfternoon: term.RGB{R: 175, G: 215, B: 215},
		MeadowTop: term.RGB{R: 95, G: 135, B: 0}, MeadowBot: term.RGB{R: 0, G: 95, B: 0}, Scrub: term.RGB{R: 135, G: 175, B: 0},
		WarmMid: term.RGB{R: 95, G: 95, B: 135}},
}

// VistaTonePick is the tone the live vista draws; 0 is today. HIS PICK,
// 2026-09-16: deep and cool (4).
var VistaTonePick = 4

// dayWeight is how much of the day a tone owns at an hour: all of it from
// 09:00 to 15:00, none at the dawn and dusk keyframes (06:00 and 18:00).
func dayWeight(tod float64) float64 {
	return math.Max(0, math.Min(1, (0.25-math.Abs(tod-0.5))/0.125))
}

// horizonAt blends a tone's three horizons by the hour.
func (t VistaTone) horizonAt(tod float64) term.RGB {
	switch {
	case tod <= 0.375:
		return t.HzMorning
	case tod <= 0.5:
		return term.Lerp(t.HzMorning, t.HzNoon, (tod-0.375)/0.125)
	case tod <= 0.625:
		return term.Lerp(t.HzNoon, t.HzAfternoon, (tod-0.5)/0.125)
	}
	return t.HzAfternoon
}

// applyTone gives the palette a tone's day sky, weighted by the hour, and
// returns the tone for the painter's meadow and its sky's middle stop.
// Tone 0 by construction returns the palette's own colours.
func applyTone(p scape.Palette, tod float64, pick int) (scape.Palette, VistaTone) {
	if pick <= 0 || pick >= len(VistaTones) {
		return p, VistaTones[0]
	}
	t := VistaTones[pick]
	w := dayWeight(tod)
	p.SkyTop = term.Lerp(p.SkyTop, t.SkyTop, w)
	p.SkyHorizon = term.Lerp(p.SkyHorizon, t.horizonAt(tod), w)
	return p, t
}

// skyMid is the colour the sky passes through halfway down: the plain
// midpoint of the two ends, pulled toward the tone's WarmMid by how warm
// the horizon is (red over blue), so a blue-to-orange dusk goes through a
// mauve and a blue-to-blue day or night does not.
func (t VistaTone) skyMid(top, hz term.RGB) term.RGB {
	mid := term.Lerp(top, hz, 0.5)
	if t.WarmMid == (term.RGB{}) {
		return mid
	}
	warmth := math.Max(0, math.Min(1, float64(int(hz.R)-int(hz.B))/120))
	return term.Lerp(mid, t.WarmMid, warmth)
}
