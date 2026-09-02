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
	//
	// MoonVis is 0.85 at every daylight keyframe rather than 0.70-0.85, and
	// that is a consequence of the cube: a sky that keeps its blue on a
	// 256-colour terminal has to be brighter than one that does not, and the
	// disc is painted INTO that sky at this alpha. Left at 0.70 the sun read
	// +50.5 luma above its sky at dawn where it used to read +64.4 -- still
	// over the floor of 40 the test enforces, but 14 points of headroom given
	// away for nothing. The disc is nearly opaque all day now and fully opaque
	// at night.
	StarVis, MoonVis float64
}

// The sky and the sea are chosen from what the xterm-256 palette actually
// holds, the same way the beach band is.
//
// This is his ruling of 2026-09-02 -- "optimize the experience for terminal.app"
// -- applied to the two biggest regions on screen. Measured before the change,
// across 48 half-hours of the day: SeaFar landed on the GREYSCALE ramp at 40 of
// them and SkyTop at 36, so for most of a working day the sea and the sky had
// no colour at all on the target terminal. That is not a gradient that bands;
// it is a picture with the colour removed. The cause is the cube's shape rather
// than the choice of blue: below luma 60 it holds nothing blue except the pure
// #0000xx column, so any dark blue loses to the grey ramp on distance.
//
// Two consequences, and both are load-bearing:
//
//   - Daylight is BRIGHTER than it was. A sea that holds its blue has to sit
//     where the cube keeps blues, and that is above luma 70. The alternative is
//     the grey it was already showing.
//   - Night is left alone. Only 7 of the 216 cube colours are both dark and
//     coloured and five are the pure-blue column, which was rendered and
//     rejected -- see Shore.BlueSky. A monochrome night is the honest answer
//     and it is already what shipped.
//
// Every value below is a cube entry, so it survives quantisation unchanged and
// a truecolor terminal shows the same picture rather than a nicer one.
var (
	// Sky. The endpoints are chosen for how many cube colours the ramp BETWEEN
	// them crosses, not for how good either one looks on its own.
	//
	// That is the whole trick and it is not the obvious move. A band appears
	// wherever the ramp crosses a quantisation boundary in ANY channel, so the
	// smoothest gradient is the one whose channels travel furthest. #005fd7 to
	// #afd7ff is the better-looking pair and renders worse -- 7 bands against
	// 10 over seventeen rows -- because its blue channel starts at 215 and
	// crosses one boundary the whole way down. Searched over every surviving
	// cube blue; these are the best pairs that still read as a sky and a sea.
	//
	// The search has to be told what "still reads as a sky" MEANS, and the
	// first version was not. Ranked on band count alone it picked #d7d7ff for
	// the horizon, which is ten bands and a lavender stripe across the middle
	// of the sky: the ramp passes through rgb(175,175,255), where red has
	// caught green up and blue is still high. Blue with no green lead is
	// violet. The constraint that fixes it -- green at least 20 above red
	// wherever blue leads -- is what TestTheSkyNeverGoesViolet holds.
	skyDeep = term.RGB{R: 0, G: 95, B: 175} // 25  zenith, mid-morning to mid-afternoon
	//
	// One zenith for the whole working day, and the horizon carries the hour.
	// It is not only that a real sky barely moves between nine and three: the
	// sun is painted INTO the upper sky at MoonVis 0.7-0.85, so every point of
	// luma spent up there comes off the context readout. A zenith of #0087d7
	// instead of this one cost the moon 8 of its 54 luma of headroom.
	skyPale = term.RGB{R: 175, G: 215, B: 255} // 153 horizon at noon
	skyDawn = term.RGB{R: 0, G: 95, B: 135}    // 24  deep blue, before the sun
	skyDusk = term.RGB{R: 0, G: 95, B: 175}    // 25  and the same blue after it
	//
	// Dusk was a violet, #5f00af, and it looked the most distinctive sky in the
	// set. It also put a GREY one in the middle of the afternoon. The leg from
	// the daylight zenith to a violet runs red up from 0 to 95 while green runs
	// down from 95 to 0, and half way along that is rgb(48,48,175) -- which has
	// plenty of chroma and no home in the cube, whose nearest dark blue-violets
	// are all further off than grey 58. Measured at 16:30, on a clear afternoon.
	//
	// So both twilights are deep blue overhead and the horizon carries the
	// hour, which is what a real one does anyway.
	// Both are dark -- luma 75 and 48 -- and both still hold their hue, which
	// the BlueSky note said was impossible. It counted the cube's dark colours
	// below luma 40 and found only the pure #0000xx column. Between 40 and 80
	// there is more: the violets at #5f00af and #5f00d7, the blues at #005faf
	// and #005fd7. That band is exactly where a twilight zenith sits, and it is
	// why dawn and dusk do not have to be grey even though midnight does.
	//
	// It is also load-bearing rather than decorative. A first pass put the dusk
	// zenith at #875faf, luma 116, and the moon test went red: the sun sits in
	// the upper sky, it is painted into the BACKGROUND at MoonVis 0.7, and a
	// bright sky eats the contrast that carries context remaining. +37.9 luma
	// against a floor of 40.
	firePeach  = term.RGB{R: 255, G: 175, B: 135} // 216 dawn horizon
	fireOrange = term.RGB{R: 255, G: 135, B: 95}  // 209 dusk horizon
	skyCool    = term.RGB{R: 175, G: 215, B: 255} // 153 horizon, mid-morning (= skyPale)
	skyWarm    = term.RGB{R: 255, G: 215, B: 175} // 223 horizon, mid-afternoon

	// Sea. The same rule, and it mattered more here: the old pair gave four
	// colours over fifteen rows with eight of them the same cell, and #005f5f
	// to #87afff gives ten with the longest run halved.
	//
	// The near end is also held by a SECOND constraint that the search does not
	// see. A longer ramp means a given movement of the mean waterline repaints
	// more cells, and the open sea holding still is a thing he asked for by
	// name. Of the pairs that gave ten bands, this is the one that also keeps
	// the backdrop under the 8% churn TestTheBackdropHoldsStillBetweenFrames
	// allows -- 7.83%. Two of the better-looking pairs are at 8.7 and 9.6.
	//
	// The sea holds one ramp from mid-morning to mid-afternoon on purpose. Time
	// of day is the SKY's channel; the sea's job is the agent's work, and a
	// backdrop that keeps shifting under it is noise in the wrong signal.
	seaDeep = term.RGB{R: 0, G: 95, B: 95} // 23  deep water, every hour --
	//                                              the darkest the cube can hold and
	//                                              still be water, since every darker
	//                                              blue is the pure #0000xx column
	//                                              Shore.BlueSky documents as rejected
	seaShallow = term.RGB{R: 135, G: 175, B: 255} // 75  shallow at noon
	seaSlate   = term.RGB{R: 95, G: 95, B: 135}   // 60  the low-light sea
	seaSteel   = term.RGB{R: 95, G: 135, B: 175}  // 67  the sea under a dusk sky
	//
	// Not the mauve it was first given. A dusk sea reflects the darkening blue
	// overhead, not the orange on the horizon -- the warm at that hour belongs
	// to the glitter, which is already #ffaf87. And the mauve broke the leg
	// into it: #5fafd7 to #875f87 crosses green from 175 down to 95, and the
	// middle of that crossing is rgb(122,122,162), which the terminal shows as
	// grey 128. Five in the evening, the sea grey, on the day the whole point
	// of the change was that it should not be.
)

// dayKeys are keyframes around the clock; everything between is interpolated.
//
// Six, not four. The two extra sit at mid-morning and mid-afternoon, and they
// are there because the interpolation itself was destroying colour: a six-hour
// leg from a warm horizon to a cool one is a straight line through RGB, and the
// middle of that line is desaturated. Measured at 09:00 the old sky ran from a
// blue-violet to a pink-grey and spent EIGHT of seventeen rows on one colour.
// Shorter legs, each between two colours the cube holds, keep the crossing
// short and put the pale part of it where a real morning has one.
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
		Star:  term.RGB{R: 206, G: 216, B: 240}, Moon: term.RGB{R: 215, G: 215, B: 255},
		Glitter: term.RGB{R: 215, G: 215, B: 255},
		StarVis: 1.0, MoonVis: 1.0,
	}},
	{0.25, Palette{ // dawn
		SkyTop: skyDawn, SkyHorizon: firePeach,
		SeaFar: seaDeep, SeaNear: seaSlate,
		Foam:    term.RGB{R: 250, G: 226, B: 210},
		WetSand: term.RGB{R: 92, G: 76, B: 68}, SandNear: term.RGB{R: 152, G: 126, B: 102},
		Grain: term.RGB{R: 190, G: 162, B: 134},
		Star:  term.RGB{R: 210, G: 218, B: 240}, Moon: term.RGB{R: 255, G: 175, B: 95},
		Glitter: term.RGB{R: 255, G: 215, B: 175},
		StarVis: 0.25, MoonVis: 0.85,
	}},
	{0.375, Palette{ // mid-morning
		SkyTop: skyDeep, SkyHorizon: skyCool,
		SeaFar: seaDeep, SeaNear: seaShallow,
		Foam:    term.RGB{R: 247, G: 239, B: 233},
		WetSand: term.RGB{R: 108, G: 90, B: 77}, SandNear: term.RGB{R: 177, G: 151, B: 122},
		Grain: term.RGB{R: 211, G: 185, B: 154},
		Star:  term.RGB{R: 215, G: 223, B: 242}, Moon: term.RGB{R: 255, G: 195, B: 115},
		Glitter: term.RGB{R: 255, G: 215, B: 175},
		StarVis: 0.125, MoonVis: 0.85,
	}},
	{0.50, Palette{ // noon
		SkyTop: skyDeep, SkyHorizon: skyPale,
		SeaFar: seaDeep, SeaNear: seaShallow,
		Foam:    term.RGB{R: 244, G: 252, B: 255},
		WetSand: term.RGB{R: 124, G: 104, B: 86}, SandNear: term.RGB{R: 202, G: 176, B: 142},
		Grain: term.RGB{R: 232, G: 208, B: 174},
		Star:  term.RGB{R: 220, G: 228, B: 244}, Moon: term.RGB{R: 255, G: 215, B: 135},
		Glitter: term.RGB{R: 255, G: 215, B: 175},
		StarVis: 0.0, MoonVis: 0.85,
	}},
	{0.625, Palette{ // mid-afternoon
		SkyTop: skyDeep, SkyHorizon: skyWarm,
		SeaFar: seaDeep, SeaNear: seaShallow,
		Foam:    term.RGB{R: 248, G: 230, B: 217},
		WetSand: term.RGB{R: 103, G: 84, B: 71}, SandNear: term.RGB{R: 171, G: 143, B: 114},
		Grain: term.RGB{R: 205, G: 176, B: 143},
		Star:  term.RGB{R: 217, G: 224, B: 243}, Moon: term.RGB{R: 255, G: 195, B: 115},
		Glitter: term.RGB{R: 255, G: 195, B: 155},
		StarVis: 0.175, MoonVis: 0.85,
	}},
	{0.75, Palette{ // dusk
		SkyTop: skyDusk, SkyHorizon: fireOrange,
		SeaFar: seaDeep, SeaNear: seaSteel,
		Foam:    term.RGB{R: 252, G: 208, B: 178},
		WetSand: term.RGB{R: 82, G: 64, B: 56}, SandNear: term.RGB{R: 140, G: 110, B: 86},
		Grain: term.RGB{R: 178, G: 144, B: 112},
		Star:  term.RGB{R: 214, G: 220, B: 242}, Moon: term.RGB{R: 255, G: 175, B: 95},
		Glitter: term.RGB{R: 255, G: 175, B: 135},
		StarVis: 0.35, MoonVis: 0.85,
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
